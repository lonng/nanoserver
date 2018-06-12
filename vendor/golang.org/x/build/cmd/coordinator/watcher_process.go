// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The watcher binary watches the specified repositories for new
// commits and reports them to the build dashboard. This binary is
// compiled in to the coordinator binary and runs in a Docker
// container (see env/watcher-world) via the coordinator.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	goBase         = "https://go.googlesource.com/"
	watcherVersion = 3        // must match dashboard/app/build/handler.go's watcherVersion
	master         = "master" // name of the master branch
	metaURL        = goBase + "?b=master&format=JSON"
)

var (
	repoURL      = flag.String("watcher.repo", goBase+"go", "Repository URL")
	dashFlag     = flag.String("watcher.dash", "https://build.golang.org/", "Dashboard URL (must end in /)")
	keyFile      = flag.String("watcher.key", defaultKeyFile, "Build dashboard key file")
	pollInterval = flag.Duration("watcher.poll", 10*time.Second, "Remote repo poll interval")
	network      = flag.Bool("watcher.network", true, "Enable network calls (disable for testing)")
	mirrorBase   = flag.String("watcher.mirror", "", `Mirror repository base URL (eg "https://github.com/golang/")`)
	filter       = flag.String("watcher.filter", "", "If non-empty, a comma-separated list of directories or files to watch for new commits (only works on main repo). If empty, watch all files in repo.")
	branches     = flag.String("watcher.branches", "", "If non-empty, a comma-separated list of branches to watch. If empty, watch changes on every branch.")
	httpAddr     = flag.String("watcher.http", "", "If non-empty, the listen address to run an HTTP server on")
	report       = flag.Bool("watcher.report", true, "Report updates to build dashboard (use false for development dry-run mode)")
)

var (
	defaultKeyFile = filepath.Join(homeDir(), ".gobuildkey")
	dashboardKey   = ""
	networkSeen    = make(map[string]bool) // track known hashes for testing
)

func watcherMain() {
	log.Printf("Running watcher role.")
	go pollGerritAndTickle()
	err := runWatcher()
	log.Printf("Watcher exiting after failure: %v", err)
	os.Exit(1)
}

// runWatcher is a little wrapper so we can use defer and return to signal
// errors. It should only return a non-nil error.
func runWatcher() error {
	if !strings.HasSuffix(*dashFlag, "/") {
		return errors.New("dashboard URL (-dashboard) must end in /")
	}

	if *report {
		if k, err := readKey(); err != nil {
			return err
		} else {
			dashboardKey = k
		}
	}

	var dir string
	if fi, err := os.Stat(watcherGitCacheDir); err == nil && fi.IsDir() {
		dir = watcherGitCacheDir
	} else {
		var err error
		dir, err = ioutil.TempDir("", "watcher")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)
	}

	if *httpAddr != "" {
		ln, err := net.Listen("tcp", *httpAddr)
		if err != nil {
			return err
		}
		go http.Serve(ln, nil)
	}

	errc := make(chan error)

	go func() {
		dst := ""
		if *mirrorBase != "" {
			name := (*repoURL)[strings.LastIndex(*repoURL, "/")+1:]
			dst = *mirrorBase + name
		}
		name := strings.TrimPrefix(*repoURL, goBase)
		r, err := NewRepo(dir, *repoURL, dst, "", true)
		if err != nil {
			errc <- err
			return
		}
		http.Handle("/"+name+".tar.gz", r)
		errc <- r.Watch()
	}()

	subrepos, err := subrepoList()
	if err != nil {
		return err
	}

	start := func(name, path string, dash bool) {
		url := goBase + name
		dst := ""
		if *mirrorBase != "" {
			dst = *mirrorBase + name
			if !repoExists(dst) {
				log.Println("skipping mirror to nonexistent repo:", dst)
				dst = ""
			}
		}
		r, err := NewRepo(dir, url, dst, path, dash)
		if err != nil {
			errc <- err
			return
		}
		http.Handle("/"+name+".tar.gz", r)
		errc <- r.Watch()
	}

	seen := map[string]bool{"go": true}
	for _, path := range subrepos {
		name := strings.TrimPrefix(path, "golang.org/x/")
		seen[name] = true
		go start(name, path, true)
	}
	if *mirrorBase != "" {
		for name := range gerritMetaMap() {
			if seen[name] {
				// Repo already picked up by dashboard list.
				continue
			}
			go start(name, "golang.org/x/"+name, false)
		}
	}

	// Must be non-nil.
	return <-errc
}

func repoExists(url string) bool {
	r, err := http.Get(url)
	if err != nil {
		log.Printf("repoExists %v: %v", url, err)
		return false
	}
	r.Body.Close()
	return r.StatusCode/100 == 2
}

// Repo represents a repository to be watched.
type Repo struct {
	root     string             // on-disk location of the git repo
	path     string             // base import path for repo (blank for main repo)
	commits  map[string]*Commit // keyed by full commit hash (40 lowercase hex digits)
	branches map[string]*Branch // keyed by branch name, eg "release-branch.go1.3" (or empty for default)
	dash     bool               // push new commits to the dashboard
	mirror   bool               // push new commits to 'dest' remote
}

// NewRepo checks out a new instance of the Mercurial repository
// specified by srcURL to a new directory inside dir.
// If dstURL is not empty, changes from the source repository will
// be mirrored to the specified destination repository.
// The importPath argument is the base import path of the repository,
// and should be empty for the main Go repo.
// The dash argument should be set true if commits to this
// repo should be reported to the build dashboard.
func NewRepo(dir, srcURL, dstURL, importPath string, dash bool) (*Repo, error) {
	var root string
	if importPath == "" {
		root = filepath.Join(dir, "go")
	} else {
		root = filepath.Join(dir, path.Base(importPath))
	}
	r := &Repo{
		path:     importPath,
		root:     root,
		commits:  make(map[string]*Commit),
		branches: make(map[string]*Branch),
		mirror:   dstURL != "",
		dash:     dash,
	}

	needClone := true
	if r.shouldTryReuseGitDir(dstURL) {
		cmd := exec.Command("git", "fetch", "--all")
		cmd.Dir = r.root
		r.logf("running git fetch --all")
		t0 := time.Now()
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			r.logf("git fetch --all failed; proceeding to wipe + clone instead; err: %v, stderr: %s", err, stderr.Bytes())
		} else {
			needClone = false
			r.logf("ran git fetch --all in %v", time.Since(t0))
		}
	}
	if needClone {
		os.RemoveAll(r.root)
		t0 := time.Now()
		r.logf("cloning %v", srcURL)
		cmd := exec.Command("git", "clone", "--mirror", srcURL, r.root)
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("cloning %s: %v\n\n%s", srcURL, err, out)
		}
		r.logf("cloned in %v", time.Since(t0))
	}

	if r.mirror {
		if err := r.addRemote("dest", dstURL); err != nil {
			return nil, fmt.Errorf("adding remote: %v", err)
		}
		r.logf("initial push to %v", dstURL)
		if err := r.push(); err != nil {
			return nil, err
		}
	}

	if r.dash {
		r.logf("loading commit log")
		if err := r.update(false); err != nil {
			return nil, err
		}
		r.logf("found %v branches among %v commits\n", len(r.branches), len(r.commits))
	}

	return r, nil
}

// shouldTryReuseGitDir reports whether we should try to reuse r.root as the git
// directory. (The directory may be corrupt, though.)
// dstURL is optional, and is the desired remote URL for a remote named "dest".
func (r *Repo) shouldTryReuseGitDir(dstURL string) bool {
	if _, err := os.Stat(filepath.Join(r.root, "FETCH_HEAD")); err != nil {
		return false
	}
	if dstURL == "" {
		return true
	}

	// Does the "dest" remote match? If not, we return false and nuke
	// the world and re-clone out of laziness.
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = r.root
	out, err := cmd.Output()
	if err != nil {
		log.Printf("git remote -v: %v", err)
	}
	for _, ln := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(ln, "dest") {
			continue
		}
		f := strings.Fields(ln)
		if len(f) < 2 {
			continue
		}
		if f[0] == "dest" && f[1] == dstURL {
			return true
		}
	}
	r.logf("not reusing old repo: remote \"dest\" URL doesn't match")
	return false
}

func (r *Repo) addRemote(name, url string) error {
	gitConfig := filepath.Join(r.root, "config")
	f, err := os.OpenFile(gitConfig, os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "\n[remote %q]\n\turl = %v\n", name, url)
	if err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// Watch continuously runs "git fetch" in the repo, checks for
// new commits, posts any new commits to the dashboard (if enabled),
// and mirrors commits to a destination repo (if enabled).
// It only returns a non-nil error.
func (r *Repo) Watch() error {
	tickler := repoTickler(r.name())
	for {
		if err := r.fetch(); err != nil {
			return err
		}
		if r.mirror {
			if err := r.push(); err != nil {
				return err
			}
		}
		if r.dash {
			if err := r.updateDashboard(); err != nil {
				return err
			}
		}
		// We still run a timer but a very slow one, just
		// in case the mechanism updating the repo tickler
		// breaks for some reason.
		timer := time.NewTimer(5 * time.Minute)
		select {
		case <-tickler:
			timer.Stop()
		case <-timer.C:
		}
	}
}

func (r *Repo) updateDashboard() error {
	if err := r.update(true); err != nil {
		return err
	}
	remotes, err := r.remotes()
	if err != nil {
		return err
	}
	for _, name := range remotes {
		b, ok := r.branches[name]
		if !ok {
			// skip branch; must be already merged
			continue
		}
		if err := r.postNewCommits(b); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) name() string {
	if r.path == "" {
		return "go"
	}
	return path.Base(r.path)
}

func (r *Repo) logf(format string, args ...interface{}) {
	log.Printf(r.name()+": "+format, args...)
}

// postNewCommits looks for unseen commits on the specified branch and
// posts them to the dashboard.
func (r *Repo) postNewCommits(b *Branch) error {
	if b.Head == b.LastSeen {
		return nil
	}
	c := b.LastSeen
	if c == nil {
		// Haven't seen anything on this branch yet:
		if b.Name == master {
			// For the master branch, bootstrap by creating a dummy
			// commit with a lone child that is the initial commit.
			c = &Commit{}
			for _, c2 := range r.commits {
				if c2.Parent == "" {
					c.children = []*Commit{c2}
					break
				}
			}
			if c.children == nil {
				return fmt.Errorf("couldn't find initial commit")
			}
		} else {
			// Find the commit that this branch forked from.
			base, err := r.mergeBase("heads/"+b.Name, master)
			if err != nil {
				return err
			}
			var ok bool
			c, ok = r.commits[base]
			if !ok {
				return fmt.Errorf("couldn't find base commit: %v", base)
			}
		}
	}
	if err := r.postChildren(b, c); err != nil {
		return err
	}
	b.LastSeen = b.Head
	return nil
}

// postChildren posts to the dashboard all descendants of the given parent.
// It ignores descendants that are not on the given branch.
func (r *Repo) postChildren(b *Branch, parent *Commit) error {
	for _, c := range parent.children {
		if c.Branch != b.Name {
			continue
		}
		if err := r.postCommit(c); err != nil {
			if strings.Contains(err.Error(), "this package already has a first commit; aborting") {
				return nil
			}
			return err
		}
	}
	for _, c := range parent.children {
		if err := r.postChildren(b, c); err != nil {
			return err
		}
	}
	return nil
}

// postCommit sends a commit to the build dashboard.
func (r *Repo) postCommit(c *Commit) error {
	if !*report {
		r.logf("dry-run mode; NOT posting commit to dashboard: %v", c)
		return nil
	}
	r.logf("sending commit to dashboard: %v", c)

	t, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", c.Date)
	if err != nil {
		return fmt.Errorf("postCommit: parsing date %q for commit %v: %v", c.Date, c, err)
	}
	dc := struct {
		PackagePath string // (empty for main repo commits)
		Hash        string
		ParentHash  string

		User   string
		Desc   string
		Time   time.Time
		Branch string

		NeedsBenchmarking bool
	}{
		PackagePath: r.path,
		Hash:        c.Hash,
		ParentHash:  c.Parent,

		User:   c.Author,
		Desc:   c.Desc,
		Time:   t,
		Branch: c.Branch,

		NeedsBenchmarking: c.NeedsBenchmarking(),
	}
	b, err := json.Marshal(dc)
	if err != nil {
		return fmt.Errorf("postCommit: marshaling request body: %v", err)
	}

	if !*network {
		if c.Parent != "" {
			if !networkSeen[c.Parent] {
				r.logf("%v: %v", c.Parent, r.commits[c.Parent])
				return fmt.Errorf("postCommit: no parent %v found on dashboard for %v", c.Parent, c)
			}
		}
		if networkSeen[c.Hash] {
			return fmt.Errorf("postCommit: already seen %v", c)
		}
		networkSeen[c.Hash] = true
		return nil
	}

	v := url.Values{"version": {fmt.Sprint(watcherVersion)}, "key": {dashboardKey}}
	u := *dashFlag + "commit?" + v.Encode()
	resp, err := http.Post(u, "text/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("postCommit: reading body: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("postCommit: status: %v\nbody: %s", resp.Status, body)
	}

	var s struct {
		Error string
	}
	if err := json.Unmarshal(body, &s); err != nil {
		return fmt.Errorf("postCommit: decoding response: %v", err)
	}
	if s.Error != "" {
		return fmt.Errorf("postCommit: error: %v", s.Error)
	}
	return nil
}

// update looks for new commits and branches,
// and updates the commits and branches maps.
func (r *Repo) update(noisy bool) error {
	remotes, err := r.remotes()
	if err != nil {
		return err
	}
	for _, name := range remotes {
		b := r.branches[name]

		// Find all unseen commits on this branch.
		revspec := "heads/" + name
		if b != nil {
			// If we know about this branch,
			// only log commits down to the known head.
			revspec = b.Head.Hash + ".." + revspec
		} else if name != master {
			// If this is an unknown non-master branch,
			// log up to where it forked from master.
			base, err := r.mergeBase(revspec, master)
			if base == "" {
				// This branch did not fork from master so we
				// don't care about it.
				delete(r.branches, name)
				log.Printf("Found independent branch %s. This branch will not be watched.", name)
				continue
			}
			if err != nil {
				return err
			}
			revspec = base + ".." + revspec
		}
		log, err := r.log("--topo-order", revspec)
		if err != nil {
			return err
		}
		if len(log) == 0 {
			// No commits to handle; carry on.
			continue
		}

		// Add unknown commits to r.commits.
		var added []*Commit
		for _, c := range log {
			// Sanity check: we shouldn't see the same commit twice.
			if dup, ok := r.commits[c.Hash]; ok {
				return fmt.Errorf("found commit we already knew about: %v; first seen on %s, now on %s", c, name, dup.Branch)
			}
			if noisy {
				r.logf("found new commit %v", c)
			}
			c.Branch = name
			r.commits[c.Hash] = c
			added = append(added, c)
		}

		// Link added commits.
		for _, c := range added {
			if c.Parent == "" {
				// This is the initial commit; no parent.
				r.logf("no parents for initial commit %v", c)
				continue
			}
			// Find parent commit.
			p, ok := r.commits[c.Parent]
			if !ok {
				return fmt.Errorf("can't find parent %q for %v", c.Parent, c)
			}
			// Link parent Commit.
			c.parent = p
			// Link child Commits.
			p.children = append(p.children, c)
		}

		// Update branch head, or add newly discovered branch.
		head := log[0]
		if b != nil {
			// Known branch; update head.
			b.Head = head
			r.logf("updated branch head: %v", b)
		} else {
			// It's a new branch; add it.
			seen, err := r.lastSeen(head.Hash)
			if err != nil {
				return err
			}
			b = &Branch{Name: name, Head: head, LastSeen: seen}
			r.branches[name] = b
			r.logf("found branch: %v", b)
		}
	}

	return nil
}

// lastSeen finds the most recent commit the dashboard has seen,
// starting at the specified head. If the dashboard hasn't seen
// any of the commits from head to the beginning, it returns nil.
func (r *Repo) lastSeen(head string) (*Commit, error) {
	h, ok := r.commits[head]
	if !ok {
		return nil, fmt.Errorf("lastSeen: can't find %q in commits", head)
	}

	var s []*Commit
	for c := h; c != nil; c = c.parent {
		s = append(s, c)
	}

	var err error
	i := sort.Search(len(s), func(i int) bool {
		if err != nil {
			return false
		}
		ok, err = r.dashSeen(s[i].Hash)
		return ok
	})
	switch {
	case err != nil:
		return nil, fmt.Errorf("lastSeen: %v", err)
	case i < len(s):
		return s[i], nil
	default:
		// Dashboard saw no commits.
		return nil, nil
	}
}

// dashSeen reports whether the build dashboard knows the specified commit.
func (r *Repo) dashSeen(hash string) (bool, error) {
	if !*network {
		return networkSeen[hash], nil
	}
	v := url.Values{"hash": {hash}, "packagePath": {r.path}}
	u := *dashFlag + "commit?" + v.Encode()
	resp, err := http.Get(u)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("status: %v", resp.Status)
	}
	var s struct {
		Error string
	}
	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		return false, err
	}
	switch s.Error {
	case "":
		// Found one.
		return true, nil
	case "Commit not found":
		// Commit not found, keep looking for earlier commits.
		return false, nil
	default:
		return false, fmt.Errorf("dashboard: %v", s.Error)
	}
}

// mergeBase returns the hash of the merge base for revspecs a and b.
func (r *Repo) mergeBase(a, b string) (string, error) {
	cmd := exec.Command("git", "merge-base", a, b)
	cmd.Dir = r.root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git merge-base %s..%s: %v", a, b, err)
	}
	return string(bytes.TrimSpace(out)), nil
}

// remotes returns a slice of remote branches known to the git repo.
// It always puts "origin/master" first.
func (r *Repo) remotes() ([]string, error) {
	if *branches != "" {
		return strings.Split(*branches, ","), nil
	}

	cmd := exec.Command("git", "branch")
	cmd.Dir = r.root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git branch: %v", err)
	}
	bs := []string{master}
	for _, b := range strings.Split(string(out), "\n") {
		b = strings.TrimPrefix(b, "* ")
		b = strings.TrimSpace(b)
		// Ignore aliases, blank lines, and master (it's already in bs).
		if b == "" || strings.Contains(b, "->") || b == master {
			continue
		}
		// Ignore pre-go1 release branches; they are just noise.
		if strings.HasPrefix(b, "release-branch.r") {
			continue
		}
		bs = append(bs, b)
	}
	return bs, nil
}

const logFormat = `--format=format:` + logBoundary + `%H
%P
%an <%ae>
%cD
%B
` + fileBoundary

const logBoundary = `_-_- magic boundary -_-_`
const fileBoundary = `_-_- file boundary -_-_`

// log runs "git log" with the supplied arguments
// and parses the output into Commit values.
func (r *Repo) log(dir string, args ...string) ([]*Commit, error) {
	args = append([]string{"log", "--date=rfc", "--name-only", "--parents", logFormat}, args...)
	if r.path == "" && *filter != "" {
		paths := strings.Split(*filter, ",")
		args = append(args, "--")
		args = append(args, paths...)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = r.root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %v: %v\n%s", strings.Join(args, " "), err, out)
	}

	// We have a commit with description that contains 0x1b byte.
	// Mercurial does not escape it, but xml.Unmarshal does not accept it.
	// TODO(adg): do we still need to scrub this? Probably.
	out = bytes.Replace(out, []byte{0x1b}, []byte{'?'}, -1)

	var cs []*Commit
	for _, text := range strings.Split(string(out), logBoundary) {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		p := strings.SplitN(text, "\n", 5)
		if len(p) != 5 {
			return nil, fmt.Errorf("git log %v: malformed commit: %q", strings.Join(args, " "), text)
		}

		// The change summary contains the change description and files
		// modified in this commit.  There is no way to directly refer
		// to the modified files in the log formatting string, so we look
		// for the file boundary after the description.
		changeSummary := p[4]
		descAndFiles := strings.SplitN(changeSummary, fileBoundary, 2)
		desc := strings.TrimSpace(descAndFiles[0])

		// For branch merges, the list of files can still be empty
		// because there are no changed files.
		files := strings.Replace(strings.TrimSpace(descAndFiles[1]), "\n", " ", -1)

		cs = append(cs, &Commit{
			Hash: p[0],
			// TODO(adg): This may break with branch merges.
			Parent: strings.Split(p[1], " ")[0],
			Author: p[2],
			Date:   p[3],
			Desc:   desc,
			Files:  files,
		})
	}
	return cs, nil
}

// fetch runs "git fetch" in the repository root.
// It tries three times, just in case it failed because of a transient error.
func (r *Repo) fetch() error {
	return try(3, func() error {
		cmd := exec.Command("git", "fetch", "origin")
		cmd.Dir = r.root
		if out, err := cmd.CombinedOutput(); err != nil {
			err = fmt.Errorf("%v\n\n%s", err, out)
			r.logf("git fetch: %v", err)
			return err
		}
		return nil
	})
}

// push runs "git push -f --mirror dest" in the repository root.
// It tries three times, just in case it failed because of a transient error.
func (r *Repo) push() error {
	return try(3, func() error {
		cmd := exec.Command("git", "push", "-f", "--mirror", "dest")
		cmd.Dir = r.root
		if out, err := cmd.CombinedOutput(); err != nil {
			err = fmt.Errorf("%v\n\n%s", err, out)
			r.logf("git push: %v", err)
			return err
		}
		return nil
	})
}

func (r *Repo) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" && req.Method != "HEAD" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	rev := req.FormValue("rev")
	if rev == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cmd := exec.Command("git", "archive", "--format=tgz", rev)
	cmd.Dir = r.root
	tgz, err := cmd.Output()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(tgz)))
	w.Header().Set("Content-Type", "application/x-compressed")
	w.Write(tgz)
}

func try(n int, fn func() error) error {
	var err error
	for tries := 0; tries < n; tries++ {
		time.Sleep(time.Duration(tries) * 5 * time.Second) // Linear back-off.
		if err = fn(); err == nil {
			break
		}
	}
	return err
}

// Branch represents a Mercurial branch.
type Branch struct {
	Name     string
	Head     *Commit
	LastSeen *Commit // the last commit posted to the dashboard
}

func (b *Branch) String() string {
	return fmt.Sprintf("%q(Head: %v LastSeen: %v)", b.Name, b.Head, b.LastSeen)
}

// Commit represents a single Git commit.
type Commit struct {
	Hash   string
	Author string
	Date   string // Format: "Mon, 2 Jan 2006 15:04:05 -0700"
	Desc   string // Plain text, first line is a short description.
	Parent string
	Branch string
	Files  string

	// For walking the graph.
	parent   *Commit
	children []*Commit
}

func (c *Commit) String() string {
	s := c.Hash
	if c.Branch != "" {
		s += fmt.Sprintf("[%v]", c.Branch)
	}
	s += fmt.Sprintf("(%q)", strings.SplitN(c.Desc, "\n", 2)[0])
	return s
}

// NeedsBenchmarking reports whether the Commit needs benchmarking.
func (c *Commit) NeedsBenchmarking() bool {
	// Do not benchmark branch commits, they are usually not interesting
	// and fall out of the trunk succession.
	if c.Branch != master {
		return false
	}
	// Do not benchmark commits that do not touch source files (e.g. CONTRIBUTORS).
	for _, f := range strings.Split(c.Files, " ") {
		if (strings.HasPrefix(f, "include") || strings.HasPrefix(f, "src")) &&
			!strings.HasSuffix(f, "_test.go") && !strings.Contains(f, "testdata") {
			return true
		}
	}
	return false
}

func homeDir() string {
	switch runtime.GOOS {
	case "plan9":
		return os.Getenv("home")
	case "windows":
		return os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	}
	return os.Getenv("HOME")
}

func readKey() (string, error) {
	c, err := ioutil.ReadFile(*keyFile)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(bytes.SplitN(c, []byte("\n"), 2)[0])), nil
}

// subrepoList fetches a list of sub-repositories from the dashboard
// and returns them as a slice of base import paths.
// Eg, []string{"golang.org/x/tools", "golang.org/x/net"}.
func subrepoList() ([]string, error) {
	if !*network {
		return nil, nil
	}

	r, err := http.Get(*dashFlag + "packages?kind=subrepo")
	if err != nil {
		return nil, fmt.Errorf("subrepo list: %v", err)
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return nil, fmt.Errorf("subrepo list: got status %v", r.Status)
	}
	var resp struct {
		Response []struct {
			Path string
		}
		Error string
	}
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, fmt.Errorf("subrepo list: %v", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("subrepo list: %v", resp.Error)
	}
	var pkgs []string
	for _, r := range resp.Response {
		pkgs = append(pkgs, r.Path)
	}
	return pkgs, nil
}

var (
	ticklerMu sync.Mutex
	ticklers  = make(map[string]chan bool)
)

// repo is the gerrit repo: e.g. "go", "net", "crypto", ...
func repoTickler(repo string) chan bool {
	ticklerMu.Lock()
	defer ticklerMu.Unlock()
	if c, ok := ticklers[repo]; ok {
		return c
	}
	c := make(chan bool, 1)
	ticklers[repo] = c
	return c
}

// pollGerritAndTickle polls Gerrit's JSON meta URL of all its URLs
// and their current branch heads.  When this sees that one has
// changed, it tickles the channel for that repo and wakes up its
// poller, if its poller is in a sleep.
func pollGerritAndTickle() {
	last := map[string]string{} // repo -> last seen hash
	for {
		for repo, hash := range gerritMetaMap() {
			if hash != last[repo] {
				last[repo] = hash
				select {
				case repoTickler(repo) <- true:
				default:
				}
			}
		}
		time.Sleep(*pollInterval)
	}
}

// gerritMetaMap returns the map from repo name (e.g. "go") to its
// latest master hash.
// The returned map is nil on any transient error.
func gerritMetaMap() map[string]string {
	res, err := http.Get(metaURL)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	defer io.Copy(ioutil.Discard, res.Body) // ensure EOF for keep-alive
	if res.StatusCode != 200 {
		return nil
	}
	var meta map[string]struct {
		Branches map[string]string
	}
	br := bufio.NewReader(res.Body)
	// For security reasons or something, this URL starts with ")]}'\n" before
	// the JSON object. So ignore that.
	// Shawn Pearce says it's guaranteed to always be just one line, ending in '\n'.
	for {
		b, err := br.ReadByte()
		if err != nil {
			return nil
		}
		if b == '\n' {
			break
		}
	}
	if err := json.NewDecoder(br).Decode(&meta); err != nil {
		log.Printf("JSON decoding error from %v: %s", metaURL, err)
		return nil
	}
	m := map[string]string{}
	for repo, v := range meta {
		if master, ok := v.Branches["master"]; ok {
			m[repo] = master
		}
	}
	return m
}
