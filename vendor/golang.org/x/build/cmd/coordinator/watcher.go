// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Code related to managing the 'watcher' child process in
// a Docker container.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"google.golang.org/cloud/compute/metadata"
)

var (
	watchers = map[string]watchConfig{} // populated at startup, keyed by repo, e.g. "https://go.googlesource.com/go"
)

type watchConfig struct {
	repo       string        // "https://go.googlesource.com/go"
	dash       string        // "https://build.golang.org/" (must end in /)
	interval   time.Duration // Polling interval
	mirrorBase string        // "https://github.com/golang/" or empty to disable mirroring
	netHost    bool          // run docker container in the host's network namespace
	httpAddr   string
}

type imageInfo struct {
	url string // of tar file

	mu      sync.Mutex
	lastMod string
}

// watcherDockerImage is the Docker container we run in. This
// "go-watcher-world" container doesn't actually contain the watcher
// binary itself; instead, the watcher binary is this coordinator
// binary, which we bind mount into the world with "docker run -v".
// That we we only need to update the Docker environment when there
// are things we need (git, etc).
const watcherDockerImage = "go-watcher-world"

var images = map[string]*imageInfo{
	watcherDockerImage: {url: "https://storage.googleapis.com/go-builder-data/docker-watcher-world.tar.gz"},
}

const gitArchiveAddr = "127.0.0.1:21536" // 21536 == keys above WATCH

func startWatchers() {
	mirrorBase := "https://github.com/golang/"
	if inStaging {
		mirrorBase = "" // don't mirror from dev cluster
	}
	addWatcher(watchConfig{
		repo:       "https://go.googlesource.com/go",
		dash:       dashBase(),
		mirrorBase: mirrorBase,
		netHost:    true,
		httpAddr:   gitArchiveAddr,
	})
	if false {
		// TODO(cmang,adg): only use one watcher or the other, depending on which build
		// coordinator is in use.
		addWatcher(watchConfig{repo: "https://go.googlesource.com/gofrontend", dash: dashBase() + "gccgo/"})
	}

	stopWatchers() // clean up before we start new ones
	for _, watcher := range watchers {
		if err := startWatching(watchers[watcher.repo]); err != nil {
			log.Printf("Error starting watcher for %s: %v", watcher.repo, err)
		}
	}
}

// Stop any previous go-watcher-world Docker tasks, so they don't
// pile up upon restarts of the coordinator.
func stopWatchers() {
	out, err := exec.Command("docker", "ps", "--no-trunc").Output()
	if err != nil {
		return
	}
	foundOld := false
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, "go-watcher-world") {
			continue
		}
		foundOld = true
		f := strings.Fields(line)
		id := f[0]
		log.Printf("killing old watcher process %s ...", id)
		err := exec.Command("docker", "rm", "-f", "-v", id).Run()
		log.Printf("killed old watcher process %s: %v", id, err)
	}
	if !foundOld {
		return
	}
	out, _ = exec.Command("docker", "ps", "--no-trunc").Output()
	if strings.Contains(string(out), "go-watcher-world") {
		log.Printf("Failed to kill previous watchers. Current containers: %s", out)
	}
}

const watcherGitCacheDir = "/var/cache/watcher-git"

// returns the part after "docker run"
func (conf watchConfig) dockerRunArgs() (args []string) {
	log.Printf("Running watcher with master key %q", masterKey())

	if err := os.MkdirAll(watcherGitCacheDir, 0755); err != nil {
		log.Fatalf("Failed to created watcher's git cache dir: %v", err)
	}

	if key := masterKey(); len(key) > 0 {
		tmpKey := "/tmp/watcher.buildkey"
		if _, err := os.Stat(tmpKey); err != nil {
			if err := ioutil.WriteFile(tmpKey, key, 0600); err != nil {
				log.Fatal(err)
			}
		}
		args = append(args, "-v", os.Args[0]+":/usr/local/bin/watcher")
		args = append(args, "-v", watcherGitCacheDir+":"+watcherGitCacheDir)
		// Images may look for .gobuildkey in / or /root, so provide both.
		// TODO(adg): fix images that look in the wrong place.
		args = append(args, "-v", tmpKey+":/.gobuildkey")
		args = append(args, "-v", tmpKey+":/root/.gobuildkey")
	}
	if conf.netHost {
		args = append(args, "--net=host")
	}
	args = append(args,
		watcherDockerImage,
		"/usr/local/bin/watcher",
		"-role=watcher",
		"-watcher.repo="+conf.repo,
		"-watcher.dash="+conf.dash,
		"-watcher.poll="+conf.interval.String(),
		"-watcher.http="+conf.httpAddr,
	)
	if conf.mirrorBase != "" {
		dst, err := url.Parse(conf.mirrorBase)
		if err != nil {
			log.Fatalf("Bad mirror destination URL: %q", conf.mirrorBase)
		}
		dst.User = url.UserPassword(mirrorCred())
		args = append(args, "-watcher.mirror="+dst.String())
	}
	return
}

func addWatcher(c watchConfig) {
	if c.repo == "" {
		c.repo = "https://go.googlesource.com/go"
	}
	if c.dash == "" {
		c.dash = "https://build.golang.org/"
	}
	if c.interval == 0 {
		c.interval = 10 * time.Second
	}
	watchers[c.repo] = c
}

func condUpdateImage(img string) error {
	ii := images[img]
	if ii == nil {
		return fmt.Errorf("image %q doesn't exist", img)
	}
	ii.mu.Lock()
	defer ii.mu.Unlock()
	u := ii.url
	if inStaging {
		u = strings.Replace(u, "go-builder-data", "dev-go-builder-data", 1)
	}
	res, err := http.Head(u)
	if err != nil {
		return fmt.Errorf("Error checking %s: %v", u, err)
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("Error checking %s: %v", u, res.Status)
	}
	if res.Header.Get("Last-Modified") == ii.lastMod {
		return nil
	}

	res, err = http.Get(u)
	if err != nil || res.StatusCode != 200 {
		return fmt.Errorf("Get after Head failed for %s: %v, %v", u, err, res)
	}
	defer res.Body.Close()

	log.Printf("Running: docker load of %s\n", u)
	cmd := exec.Command("docker", "load")
	cmd.Stdin = res.Body

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if cmd.Run(); err != nil {
		log.Printf("Failed to pull latest %s from %s and pipe into docker load: %v, %s", img, u, err, out.Bytes())
		return err
	}
	ii.lastMod = res.Header.Get("Last-Modified")
	return nil
}

var (
	watchLogMu     sync.Mutex
	watchLastFail  = map[string]string{} // repo -> logs
	watchContainer = map[string]string{} // repo -> container
)

var matchTokens = regexp.MustCompile(`\b[0-9a-f]{40}\b`)

func handleDebugWatcher(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	watchLogMu.Lock()
	defer watchLogMu.Unlock()
	for repo, logs := range watchLastFail {
		fmt.Fprintf(w, "============== Watcher %s, last fail:\n%s\n\n", repo, matchTokens.ReplaceAllString(logs, "---40hexomitted---"))
	}
	for repo, container := range watchContainer {
		logs, _ := exec.Command("docker", "logs", container).CombinedOutput()
		fmt.Fprintf(w, "============== Watcher %s, current container logs:\n%s\n\n", repo, matchTokens.ReplaceAll(logs, []byte("---40hexomitted---")))
	}
}

func startWatching(conf watchConfig) (err error) {
	defer func() {
		if err != nil {
			restartWatcherSoon(conf)
		}
	}()
	log.Printf("Starting watcher for %v", conf.repo)
	if err := condUpdateImage(watcherDockerImage); err != nil {
		log.Printf("Failed to setup container for commit watcher: %v", err)
		return err
	}

	cmd := exec.Command("docker", append([]string{"run", "-d"}, conf.dockerRunArgs()...)...)
	all, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Docker run for commit watcher = err:%v, output: %s", err, all)
		return err
	}

	container := strings.TrimSpace(string(all))

	watchLogMu.Lock()
	watchContainer[conf.repo] = container
	watchLogMu.Unlock()

	// Start a goroutine to wait for the watcher to die.
	go func() {
		exec.Command("docker", "wait", container).Run()
		out, _ := exec.Command("docker", "logs", container).CombinedOutput()
		exec.Command("docker", "rm", "-v", container).Run()
		const maxLogBytes = 512 << 10
		if len(out) > maxLogBytes {
			var partial bytes.Buffer
			partial.Write(out[:maxLogBytes/2])
			partial.WriteString("\n...(omitted)...\n")
			partial.Write(out[len(out)-(maxLogBytes/2):])
			out = partial.Bytes()
		}
		watchLogMu.Lock()
		watchLastFail[conf.repo] = string(out)
		watchLogMu.Unlock()
		log.Printf("Watcher %v crashed. Restarting soon. Logs: %s", conf.repo, out)
		restartWatcherSoon(conf)
	}()
	return nil
}

func restartWatcherSoon(conf watchConfig) {
	time.AfterFunc(30*time.Second, func() {
		startWatching(conf)
	})
}

func mirrorCred() (username, password string) {
	mirrorCredOnce.Do(loadMirrorCred)
	return mirrorCredCache.username, mirrorCredCache.password
}

var (
	mirrorCredOnce  sync.Once
	mirrorCredCache struct {
		username, password string
	}
)

func loadMirrorCred() {
	cred, err := metadata.ProjectAttributeValue("mirror-credentials")
	if err != nil {
		log.Printf("No mirror credentials available: %v", err)
		return
	}
	p := strings.SplitN(strings.TrimSpace(cred), ":", 2)
	if len(p) != 2 {
		log.Fatalf("Bad mirror credentials: %q", cred)
	}
	mirrorCredCache.username, mirrorCredCache.password = p[0], p[1]
}
