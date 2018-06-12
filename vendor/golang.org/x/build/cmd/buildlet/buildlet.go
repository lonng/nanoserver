// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The buildlet is an HTTP server that untars content to disk and runs
// commands it has untarred, streaming their output back over HTTP.
// It is part of Go's continuous build system.
//
// This program intentionally allows remote code execution, and
// provides no security of its own. It is assumed that any user uses
// it with an appropriately-configured firewall between their VM
// instances.
package main // import "golang.org/x/build/cmd/buildlet"

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha1"
	"crypto/tls"
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
	"strconv"
	"strings"
	"time"

	"golang.org/x/build/buildlet"
	"golang.org/x/build/envutil"
	"golang.org/x/build/internal/httpdl"
	"golang.org/x/build/pargzip"
	"google.golang.org/cloud/compute/metadata"
)

var (
	haltEntireOS = flag.Bool("halt", true, "halt OS in /halt handler. If false, the buildlet process just ends.")
	workDir      = flag.String("workdir", "", "Temporary directory to use. The contents of this directory may be deleted at any time. If empty, TempDir is used to create one.")
	listenAddr   = flag.String("listen", "AUTO", "address to listen on. Unused in reverse mode. Warning: this service is inherently insecure and offers no protection of its own. Do not expose this port to the world.")
	reverse      = flag.String("reverse", "", "if non-empty, go into reverse mode where the buildlet dials the coordinator instead of listening for connections. The value is a comma-separated list of modes, e.g. 'darwin-arm,darwin-amd64-race'")
	coordinator  = flag.String("coordinator", "localhost:8119", "address of coordinator, in production use farmer.golang.org. Only used in reverse mode.")
	hostname     = flag.String("hostname", "", "hostname to advertise to coordinator for reverse mode; default is actual hostname")
)

// Bump this whenever something notable happens, or when another
// component needs a certain feature. This shows on the coordinator
// per reverse client, and is also accessible via the buildlet
// package's client API (via the Status method).
//
// Notable versions:
//    3: switched to revdial protocol
//    5: reverse dialing uses timeouts+tcp keepalives, pargzip fix
//    7: version bumps while debugging revdial hang (Issue 12816)
const buildletVersion = 7

func defaultListenAddr() string {
	if runtime.GOOS == "darwin" {
		// Darwin will never run on GCE, so let's always
		// listen on a high port (so we don't need to be
		// root).
		return ":5936"
	}
	if !metadata.OnGCE() {
		return "localhost:5936"
	}
	// In production, default to port 80 or 443, depending on
	// whether TLS is configured.
	if metadataValue("tls-cert") != "" {
		return ":443"
	}
	return ":80"
}

var osHalt func() // set by some machines

func main() {
	if runtime.GOOS == "plan9" {
		log.SetOutput(&plan9LogWriter{w: os.Stderr})
	}
	onGCE := metadata.OnGCE()
	if runtime.GOOS == "linux" && onGCE && !inKube {
		if w, err := os.OpenFile("/dev/console", os.O_WRONLY, 0); err == nil {
			log.SetOutput(w)
		}
	}
	log.Printf("buildlet starting.")
	flag.Parse()

	if *listenAddr == "AUTO" {
		v := defaultListenAddr()
		log.Printf("Will listen on %s", v)
		*listenAddr = v
	}

	if !onGCE && !strings.HasPrefix(*listenAddr, "localhost:") {
		log.Printf("** WARNING ***  This server is unsafe and offers no security. Be careful.")
	}
	if onGCE {
		fixMTU()
	}
	if *workDir == "" {
		switch runtime.GOOS {
		case "windows":
			// We want a short path on Windows, due to
			// Windows issues with maximum path lengths.
			*workDir = `C:\workdir`
			if err := os.MkdirAll(*workDir, 0755); err != nil {
				log.Fatalf("error creating workdir: %v", err)
			}
		default:
			dir := filepath.Join(os.TempDir(), "workdir")
			if err := os.RemoveAll(dir); err != nil { // should be no-op
				log.Fatal(err)
			}
			if err := os.Mkdir(dir, 0755); err != nil {
				log.Fatal(err)
			}
			*workDir = dir
		}
	}
	// This is hard-coded because the client-supplied environment has
	// no way to expand relative paths from the workDir.
	// TODO(bradfitz): if we ever need more than this, make some mechanism.
	os.Setenv("GOROOT_BOOTSTRAP", filepath.Join(*workDir, "go1.4"))
	os.Setenv("WORKDIR", *workDir) // mostly for demos

	if _, err := os.Lstat(*workDir); err != nil {
		log.Fatalf("invalid --workdir %q: %v", *workDir, err)
	}
	if runtime.GOOS == "solaris" && runtime.GOARCH == "amd64" {
		os.Setenv("PATH", os.Getenv("PATH")+":/opt/local/bin")
		downloadBootstrapGoroot("/root/go-solaris-amd64-bootstrap", "https://storage.googleapis.com/go-builder-data/gobootstrap-solaris-amd64.tar.gz")
	}
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/debug/goroutines", handleGoroutines)
	http.HandleFunc("/debug/x", handleX)

	var password string
	if *reverse == "" {
		password = metadataValue("password")
	}
	requireAuth := func(handler func(w http.ResponseWriter, r *http.Request)) http.Handler {
		return requirePasswordHandler{http.HandlerFunc(handler), password}
	}
	http.Handle("/writetgz", requireAuth(handleWriteTGZ))
	http.Handle("/write", requireAuth(handleWrite))
	http.Handle("/exec", requireAuth(handleExec))
	http.Handle("/halt", requireAuth(handleHalt))
	http.Handle("/tgz", requireAuth(handleGetTGZ))
	http.Handle("/removeall", requireAuth(handleRemoveAll))
	http.Handle("/workdir", requireAuth(handleWorkDir))
	http.Handle("/status", requireAuth(handleStatus))
	http.Handle("/ls", requireAuth(handleLs))

	if *reverse == "" {
		listenForCoordinator()
	} else {
		log.Fatal(dialCoordinator())
	}
}

func downloadBootstrapGoroot(destDir, url string) {
	tarPath := destDir + ".tar.gz"
	origInfo, err := os.Stat(tarPath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Checking for tar existence: %v", err)
	}
	if err := httpdl.Download(tarPath, url); err != nil {
		log.Fatalf("Downloading %s to %s: %v", url, tarPath, err)
	}
	newInfo, err := os.Stat(tarPath)
	if err != nil {
		log.Fatalf("Stat after download: %v", err)
	}
	if os.SameFile(origInfo, newInfo) {
		// The file on disk was unmodified, so we probably untarred it already.
		return
	}
	f, err := os.Open(tarPath)
	if err != nil {
		log.Fatalf("Opening after download: %v", err)
	}
	defer f.Close()
	if err := untar(f, destDir); err != nil {
		os.Remove(tarPath)
		os.RemoveAll(destDir)
		log.Fatalf("Untarring %s: %v", url, err)
	}
}

func listenForCoordinator() {
	tlsCert, tlsKey := metadataValue("tls-cert"), metadataValue("tls-key")
	if (tlsCert == "") != (tlsKey == "") {
		log.Fatalf("tls-cert and tls-key must both be supplied, or neither.")
	}

	log.Printf("Listening on %s ...", *listenAddr)
	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", *listenAddr, err)
	}
	ln = tcpKeepAliveListener{ln.(*net.TCPListener)}

	var srv http.Server
	if tlsCert != "" {
		cert, err := tls.X509KeyPair([]byte(tlsCert), []byte(tlsKey))
		if err != nil {
			log.Fatalf("TLS cert error: %v", err)
		}
		tlsConf := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		ln = tls.NewListener(ln, tlsConf)
	}

	log.Fatalf("Serve: %v", srv.Serve(ln))
}

var inKube, _ = strconv.ParseBool(os.Getenv("IN_KUBERNETES"))

// metadataValue returns the GCE metadata instance value for the given key.
// If the metadata is not defined, the returned string is empty.
//
// If not running on GCE, it falls back to using environment variables
// for local development.
func metadataValue(key string) string {

	// The common case (on GCE, but not in Kubernetes):
	if metadata.OnGCE() && !inKube {
		v, err := metadata.InstanceAttributeValue(key)
		if _, notDefined := err.(metadata.NotDefinedError); notDefined {
			return ""
		}
		if err != nil {
			log.Fatalf("metadata.InstanceAttributeValue(%q): %v", key, err)
		}
		return v
	}

	// Else allow use of environment variables to fake
	// metadata keys, for Kubernetes pods or local testing.
	envKey := "META_" + strings.Replace(key, "-", "_", -1)
	v := os.Getenv(envKey)
	// Respect curl-style '@' prefix to mean the rest is a filename.
	if strings.HasPrefix(v, "@") {
		slurp, err := ioutil.ReadFile(v[1:])
		if err != nil {
			log.Fatalf("Error reading file for GCEMETA_%v: %v", key, err)
		}
		return string(slurp)
	}
	if v == "" {
		log.Printf("Warning: not running on GCE, and no %v environment variable defined", envKey)
	}
	return v
}

// tcpKeepAliveListener is a net.Listener that sets TCP keep-alive
// timeouts on accepted connections.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func fixMTU_freebsd() error { return fixMTU_ifconfig("vtnet0") }
func fixMTU_openbsd() error { return fixMTU_ifconfig("vio0") }
func fixMTU_ifconfig(iface string) error {
	out, err := exec.Command("/sbin/ifconfig", iface, "mtu", "1460").CombinedOutput()
	if err != nil {
		return fmt.Errorf("/sbin/ifconfig %s mtu 1460: %v, %s", iface, err, out)
	}
	return nil
}

func fixMTU_plan9() error {
	f, err := os.OpenFile("/net/ipifc/0/ctl", os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(f, "mtu 1460\n"); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func fixMTU() {
	fn, ok := map[string]func() error{
		"openbsd": fixMTU_openbsd,
		"freebsd": fixMTU_freebsd,
		"plan9":   fixMTU_plan9,
	}[runtime.GOOS]
	if ok {
		if err := fn(); err != nil {
			log.Printf("Failed to set MTU: %v", err)
		} else {
			log.Printf("Adjusted MTU.")
		}
	}
}

// flushWriter is an io.Writer that Flushes after each Write if the
// underlying Writer implements http.Flusher.
type flushWriter struct {
	rw http.ResponseWriter
}

func (fw flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.rw.Write(p)
	if f, ok := fw.rw.(http.Flusher); ok {
		f.Flush()
	}
	return
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "buildlet running on %s-%s\n", runtime.GOOS, runtime.GOARCH)
}

// unauthenticated /debug/goroutines handler
func handleGoroutines(w http.ResponseWriter, r *http.Request) {
	log.Printf("Dumping goroutines.")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	buf := make([]byte, 2<<20)
	buf = buf[:runtime.Stack(buf, true)]
	w.Write(buf)
	log.Printf("Dumped goroutines.")
}

// unauthenticated /debug/x handler, to test MTU settings.
func handleX(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.FormValue("n"))
	if n > 1<<20 {
		n = 1 << 20
	}
	log.Printf("Dumping %d X.", n)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = 'X'
	}
	w.Write(buf)
	log.Printf("Dumped X.")
}

// This is a remote code execution daemon, so security is kinda pointless, but:
func validRelativeDir(dir string) bool {
	if strings.Contains(dir, `\`) || path.IsAbs(dir) {
		return false
	}
	dir = path.Clean(dir)
	if strings.HasPrefix(dir, "../") || strings.HasSuffix(dir, "/..") || dir == ".." {
		return false
	}
	return true
}

func handleGetTGZ(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "requires GET method", http.StatusBadRequest)
		return
	}
	dir := r.FormValue("dir")
	if !validRelativeDir(dir) {
		http.Error(w, "bogus dir", http.StatusBadRequest)
		return
	}
	zw := pargzip.NewWriter(w)
	tw := tar.NewWriter(zw)
	base := filepath.Join(*workDir, filepath.FromSlash(dir))
	err := filepath.Walk(base, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(filepath.ToSlash(strings.TrimPrefix(path, base)), "/")
		var linkName string
		if fi.Mode()&os.ModeSymlink != 0 {
			linkName, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}
		th, err := tar.FileInfoHeader(fi, linkName)
		if err != nil {
			return err
		}
		th.Name = rel
		if fi.IsDir() && !strings.HasSuffix(th.Name, "/") {
			th.Name += "/"
		}
		if th.Name == "/" {
			return nil
		}
		if err := tw.WriteHeader(th); err != nil {
			return err
		}
		if fi.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Walk error: %v", err)
		// Decent way to signal failure to the caller, since it'll break
		// the chunked response, rather than have a valid EOF.
		conn, _, _ := w.(http.Hijacker).Hijack()
		conn.Close()
	}
	tw.Close()
	zw.Close()
}

func handleWriteTGZ(w http.ResponseWriter, r *http.Request) {
	urlParam, _ := url.ParseQuery(r.URL.RawQuery)
	baseDir := *workDir
	if dir := urlParam.Get("dir"); dir != "" {
		if !validRelativeDir(dir) {
			http.Error(w, "bogus dir", http.StatusBadRequest)
			return
		}
		dir = filepath.FromSlash(dir)
		baseDir = filepath.Join(baseDir, dir)

		// Special case: if the directory is "go1.4" and it already exists, do nothing.
		// This lets clients do a blind write to it and not do extra work.
		if r.Method == "POST" && dir == "go1.4" {
			if fi, err := os.Stat(baseDir); err == nil && fi.IsDir() {
				log.Printf("skipping URL puttar to go1.4 dir; already exists")
				io.WriteString(w, "SKIP")
				return
			}
		}

		if err := os.MkdirAll(baseDir, 0755); err != nil {
			http.Error(w, "mkdir of base: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var tgz io.Reader
	switch r.Method {
	case "PUT":
		tgz = r.Body
	case "POST":
		urlStr := r.FormValue("url")
		if urlStr == "" {
			http.Error(w, "missing url POST param", http.StatusBadRequest)
			return
		}
		res, err := http.Get(urlStr)
		if err != nil {
			log.Printf("Failed to fetch tgz URL %s: %v", urlStr, err)
			http.Error(w, fmt.Sprintf("fetching URL %s: %v", urlStr, err), http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			log.Printf("Failed to fetch tgz URL %s: status=%v", urlStr, res.Status)
			http.Error(w, fmt.Sprintf("fetching provided url: %s", res.Status), http.StatusInternalServerError)
			return
		}
		tgz = res.Body
	default:
		http.Error(w, "requires PUT or POST method", http.StatusBadRequest)
		return
	}

	err := untar(tgz, baseDir)
	if err != nil {
		log.Printf("untar failure: %v", err)
		status := http.StatusInternalServerError
		if he, ok := err.(httpStatuser); ok {
			status = he.httpStatus()
		}
		http.Error(w, err.Error(), status)
		return
	}
	io.WriteString(w, "OK")
}

func handleWrite(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "requires POST method", http.StatusBadRequest)
		return
	}

	param, _ := url.ParseQuery(r.URL.RawQuery)

	path := param.Get("path")
	if path == "" || !validRelPath(path) {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	path = filepath.FromSlash(path)
	path = filepath.Join(*workDir, path)

	modeInt, err := strconv.ParseInt(param.Get("mode"), 10, 64)
	mode := os.FileMode(modeInt)
	if err != nil || !mode.IsRegular() {
		http.Error(w, "bad mode", http.StatusBadRequest)
		return
	}

	// Make the directory if it doesn't exist.
	// TODO(adg): support dirmode parameter?
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := writeFile(r.Body, path, mode); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.WriteString(w, "OK")
}

func writeFile(r io.Reader, path string, mode os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return err
	}
	// Try to set the mode again, in case the file already existed.
	if runtime.GOOS != "windows" {
		if err := f.Chmod(mode); err != nil {
			f.Close()
			return err
		}
	}
	return f.Close()
}

// untar reads the gzip-compressed tar file from r and writes it into dir.
func untar(r io.Reader, dir string) error {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return badRequest("requires gzip-compressed body: " + err.Error())
	}
	tr := tar.NewReader(zr)
	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("tar reading error: %v", err)
			return badRequest("tar error: " + err.Error())
		}
		if !validRelPath(f.Name) {
			return badRequest(fmt.Sprintf("tar file contained invalid name %q", f.Name))
		}
		rel := filepath.FromSlash(f.Name)
		abs := filepath.Join(dir, rel)

		fi := f.FileInfo()
		mode := fi.Mode()
		switch {
		case mode.IsRegular():
			// Make the directory. This is redundant because it should
			// already be made by a directory entry in the tar
			// beforehand. Thus, don't check for errors; the next
			// write will fail with the same error.
			os.MkdirAll(filepath.Dir(abs), 0755)
			wf, err := os.OpenFile(abs, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return err
			}
			n, err := io.Copy(wf, tr)
			if closeErr := wf.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				return fmt.Errorf("error writing to %s: %v", abs, err)
			}
			if n != f.Size {
				return fmt.Errorf("only wrote %d bytes to %s; expected %d", n, abs, f.Size)
			}
			log.Printf("wrote %s", abs)
			if !f.ModTime.IsZero() {
				if err := os.Chtimes(abs, f.ModTime, f.ModTime); err != nil {
					// benign error. Gerrit doesn't even set the
					// modtime in these, and we don't end up relying
					// on it anywhere (the gomote push command relies
					// on digests only), so this is a little pointless
					// for now.
					log.Printf("error changing modtime: %v", err)
				}
			}
		case mode.IsDir():
			if err := os.MkdirAll(abs, 0755); err != nil {
				return err
			}
		default:
			return badRequest(fmt.Sprintf("tar file entry %s contained unsupported file type %v", f.Name, mode))
		}
	}
	return nil
}

// Process-State is an HTTP Trailer set in the /exec handler to "ok"
// on success, or os.ProcessState.String() on failure.
const hdrProcessState = "Process-State"

func handleExec(w http.ResponseWriter, r *http.Request) {
	cn := w.(http.CloseNotifier)
	clientGone := cn.CloseNotify()
	handlerDone := make(chan bool)
	defer close(handlerDone)

	if r.Method != "POST" {
		http.Error(w, "requires POST method", http.StatusBadRequest)
		return
	}
	if r.ProtoMajor*10+r.ProtoMinor < 11 {
		// We need trailers, only available in HTTP/1.1 or HTTP/2.
		http.Error(w, "HTTP/1.1 or higher required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Trailer", hdrProcessState) // declare it so we can set it

	cmdPath := r.FormValue("cmd") // required
	absCmd := cmdPath
	dir := r.FormValue("dir") // optional
	sysMode := r.FormValue("mode") == "sys"
	debug, _ := strconv.ParseBool(r.FormValue("debug"))

	if sysMode {
		if cmdPath == "" {
			http.Error(w, "requires 'cmd' parameter", http.StatusBadRequest)
			return
		}
		if dir == "" {
			dir = *workDir
		} else {
			dir = filepath.FromSlash(dir)
			if !filepath.IsAbs(dir) {
				dir = filepath.Join(*workDir, dir)
			}
		}
	} else {
		if !validRelPath(cmdPath) {
			http.Error(w, "requires 'cmd' parameter", http.StatusBadRequest)
			return
		}
		absCmd = filepath.Join(*workDir, filepath.FromSlash(cmdPath))
		if dir == "" {
			dir = filepath.Dir(absCmd)
		} else {
			if !validRelPath(dir) {
				http.Error(w, "bogus 'dir' parameter", http.StatusBadRequest)
				return
			}
			dir = filepath.Join(*workDir, filepath.FromSlash(dir))
		}
	}

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	env := append(baseEnv(), r.PostForm["env"]...)
	env = envutil.Dedup(runtime.GOOS == "windows", env)
	env = setPathEnv(env, r.PostForm["path"], *workDir)

	cmd := exec.Command(absCmd, r.PostForm["cmdArg"]...)
	cmd.Dir = dir
	cmdOutput := flushWriter{w}
	cmd.Stdout = cmdOutput
	cmd.Stderr = cmdOutput
	cmd.Env = env

	log.Printf("[%p] Running %s with args %q and env %q in dir %s",
		cmd, cmd.Path, cmd.Args, cmd.Env, cmd.Dir)

	if debug {
		fmt.Fprintf(cmdOutput, ":: Running %s with args %q and env %q in dir %s\n\n",
			cmd.Path, cmd.Args, cmd.Env, cmd.Dir)
	}

	t0 := time.Now()
	err := cmd.Start()
	if err == nil {
		go func() {
			select {
			case <-clientGone:
				err := killProcessTree(cmd.Process)
				if err != nil {
					log.Printf("Kill failed: %v", err)
				}
			case <-handlerDone:
				return
			}
		}()
		err = cmd.Wait()
	}
	state := "ok"
	if err != nil {
		if ps := cmd.ProcessState; ps != nil {
			state = ps.String()
		} else {
			state = err.Error()
		}
	}
	w.Header().Set(hdrProcessState, state)
	log.Printf("[%p] Run = %s, after %v", cmd, state, time.Since(t0))
}

// setPathEnv returns a copy of the provided environment with any existing
// PATH variables replaced by the user-provided path.
// These substitutions are applied to user-supplied path elements:
//   - the string "$PATH" expands to the original PATH elements
//   - the substring "$WORKDIR" expands to the provided workDir
// A path of just ["$EMPTY"] removes the PATH variable from the environment.
func setPathEnv(env, path []string, workDir string) []string {
	if len(path) == 0 {
		return env
	}

	var (
		pathIdx  = -1
		pathOrig = ""
	)

	for i, s := range env {
		if isPathEnvPair(s) {
			pathIdx = i
			pathOrig = s[len("PaTh="):] // in whatever case
			break
		}
	}
	if len(path) == 1 && path[0] == "$EMPTY" {
		// Remove existing path variable if it exists.
		if pathIdx >= 0 {
			env = append(env[:pathIdx], env[pathIdx+1:]...)
		}
		return env
	}

	// Apply substitions to a copy of the path argument.
	path = append([]string{}, path...)
	for i, s := range path {
		if s == "$PATH" {
			path[i] = pathOrig // ok if empty
		} else {
			path[i] = strings.Replace(s, "$WORKDIR", workDir, -1)
		}
	}

	// Put the new PATH in env.
	env = append([]string{}, env...)
	pathEnv := pathEnvVar() + "=" + strings.Join(path, pathSeparator())
	if pathIdx >= 0 {
		env[pathIdx] = pathEnv
	} else {
		env = append(env, pathEnv)
	}

	return env
}

// isPathEnvPair reports whether the key=value pair s represents
// the operating system's path variable.
func isPathEnvPair(s string) bool {
	// On Unix it's PATH.
	// On Plan 9 it's path.
	// On Windows it's pAtH case-insensitive.
	if runtime.GOOS == "windows" {
		return len(s) >= 5 && strings.EqualFold(s[:5], "PATH=")
	}
	if runtime.GOOS == "plan9" {
		return strings.HasPrefix(s, "path=")
	}
	return strings.HasPrefix(s, "PATH=")
}

// On Unix it's PATH.
// On Plan 9 it's path.
// On Windows it's pAtH case-insensitive.
func pathEnvVar() string {
	if runtime.GOOS == "plan9" {
		return "path"
	}
	return "PATH"
}

func pathSeparator() string {
	if runtime.GOOS == "plan9" {
		return "\x00"
	} else {
		return string(filepath.ListSeparator)
	}
}

func baseEnv() []string {
	if runtime.GOOS == "windows" {
		return windowsBaseEnv()
	}
	return os.Environ()
}

func windowsBaseEnv() (e []string) {
	e = append(e, "GOBUILDEXIT=1") // exit all.bat with completion status
	btype, err := metadata.InstanceAttributeValue("builder-type")
	if err != nil {
		log.Fatalf("Failed to get builder-type: %v", err)
		return nil
	}
	is64 := strings.HasPrefix(btype, "windows-amd64")
	for _, pair := range os.Environ() {
		const pathEq = "PATH="
		if hasPrefixFold(pair, pathEq) {
			e = append(e, "PATH="+windowsPath(pair[len(pathEq):], is64))
		} else {
			e = append(e, pair)
		}
	}
	return e
}

// hasPrefixFold is a case-insensitive strings.HasPrefix.
func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

// windowsPath cleans the windows %PATH% environment.
// is64Bit is whether this is a windows-amd64-* builder.
// The PATH is assumed to be that of the image described in env/windows/README.
func windowsPath(old string, is64Bit bool) string {
	vv := filepath.SplitList(old)
	newPath := make([]string, 0, len(vv))
	for _, v := range vv {
		// The base VM image has both the 32-bit and 64-bit gcc installed.
		// They're both in the environment, so scrub the one
		// we don't want (TDM-GCC-64 or TDM-GCC-32).
		if strings.Contains(v, "TDM-GCC-") {
			gcc64 := strings.Contains(v, "TDM-GCC-64")
			if is64Bit != gcc64 {
				continue
			}
		}
		newPath = append(newPath, v)
	}
	return strings.Join(newPath, string(filepath.ListSeparator))
}

func handleHalt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "requires POST method", http.StatusBadRequest)
		return
	}
	log.Printf("Halting in 1 second.")
	// do the halt in 1 second, to give the HTTP response time to complete:
	time.AfterFunc(1*time.Second, haltMachine)
}

func haltMachine() {
	if !*haltEntireOS {
		log.Printf("Ending buildlet process due to halt.")
		os.Exit(0)
		return
	}
	log.Printf("Halting machine.")
	time.AfterFunc(5*time.Second, func() { os.Exit(0) })
	if osHalt != nil {
		// TODO: Windows: http://msdn.microsoft.com/en-us/library/windows/desktop/aa376868%28v=vs.85%29.aspx
		osHalt()
		os.Exit(0)
	}
	// Backup mechanism, if exec hangs for any reason:
	var err error
	switch runtime.GOOS {
	case "openbsd":
		// Quick, no fs flush, and power down:
		err = exec.Command("halt", "-q", "-n", "-p").Run()
	case "freebsd":
		// Power off (-p), via halt (-o), now.
		err = exec.Command("shutdown", "-p", "-o", "now").Run()
	case "linux":
		// Don't sync (-n), force without shutdown (-f), and power off (-p).
		err = exec.Command("/bin/halt", "-n", "-f", "-p").Run()
	case "plan9":
		err = exec.Command("fshalt").Run()
	default:
		err = errors.New("No system-specific halt command run; will just end buildlet process.")
	}
	log.Printf("Shutdown: %v", err)
	log.Printf("Ending buildlet process post-halt")
	os.Exit(0)
}

func handleRemoveAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "requires POST method", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	paths := r.Form["path"]
	if len(paths) == 0 {
		http.Error(w, "requires 'path' parameter", http.StatusBadRequest)
		return
	}
	for _, p := range paths {
		if !validRelPath(p) {
			http.Error(w, fmt.Sprintf("bad 'path' parameter: %q", p), http.StatusBadRequest)
			return
		}
	}
	for _, p := range paths {
		log.Printf("Removing %s", p)
		fullDir := filepath.Join(*workDir, filepath.FromSlash(p))
		err := os.RemoveAll(fullDir)
		if p == "." && err != nil {
			// If workDir is a mountpoint and/or contains a binary
			// using it, we can get a "Device or resource busy" error.
			// See if it's now empty and ignore the error.
			if f, oerr := os.Open(*workDir); oerr == nil {
				if all, derr := f.Readdirnames(-1); derr == nil && len(all) == 0 {
					log.Printf("Ignoring fail of RemoveAll(.)")
					err = nil
				} else {
					log.Printf("Readdir = %q, %v", all, derr)
				}
				f.Close()
			} else {
				log.Printf("Failed to open workdir: %v", oerr)
			}
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// If we nuked the work directory, recreate it.
	if err := os.MkdirAll(*workDir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleWorkDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "requires GET method", http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, *workDir)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "requires GET method", http.StatusBadRequest)
		return
	}
	status := buildlet.Status{
		Version: buildletVersion,
	}
	b, err := json.Marshal(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(b)
}

func handleLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "requires GET method", http.StatusBadRequest)
		return
	}
	dir := r.FormValue("dir")
	recursive, _ := strconv.ParseBool(r.FormValue("recursive"))
	digest, _ := strconv.ParseBool(r.FormValue("digest"))
	skip := r.Form["skip"] // '/'-separated relative dirs

	if !validRelativeDir(dir) {
		http.Error(w, "bogus dir", http.StatusBadRequest)
		return
	}
	base := filepath.Join(*workDir, filepath.FromSlash(dir))
	anyOutput := false
	err := filepath.Walk(base, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(filepath.ToSlash(strings.TrimPrefix(path, base)), "/")
		if rel == "" && fi.IsDir() {
			return nil
		}
		if fi.IsDir() {
			for _, v := range skip {
				if rel == v {
					return filepath.SkipDir
				}
			}
		}
		anyOutput = true
		fmt.Fprintf(w, "%s\t%s", fi.Mode(), rel)
		if fi.Mode().IsRegular() {
			fmt.Fprintf(w, "\t%d\t%s", fi.Size(), fi.ModTime().UTC().Format(time.RFC3339))
			if digest {
				if sha1, err := fileSHA1(path); err != nil {
					return err
				} else {
					io.WriteString(w, "\t"+sha1)
				}
			}
		} else if fi.Mode().IsDir() {
			io.WriteString(w, "/")
		}
		io.WriteString(w, "\n")
		if fi.IsDir() && !recursive {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		log.Printf("Walk error: %v", err)
		if anyOutput {
			// Decent way to signal failure to the caller, since it'll break
			// the chunked response, rather than have a valid EOF.
			conn, _, _ := w.(http.Hijacker).Hijack()
			conn.Close()
			return
		}
		http.Error(w, "Walk error: "+err.Error(), 500)
		return
	}
}

func fileSHA1(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	s1 := sha1.New()
	if _, err := io.Copy(s1, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", s1.Sum(nil)), nil
}

func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}

type httpStatuser interface {
	error
	httpStatus() int
}

type httpError struct {
	statusCode int
	msg        string
}

func (he httpError) Error() string   { return he.msg }
func (he httpError) httpStatus() int { return he.statusCode }

func badRequest(msg string) error {
	return httpError{http.StatusBadRequest, msg}
}

// requirePassword is an http.Handler auth wrapper that enforces a
// HTTP Basic password. The username is ignored.
type requirePasswordHandler struct {
	h        http.Handler
	password string // empty means no password
}

func (h requirePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, gotPass, _ := r.BasicAuth()
	if h.password != "" && h.password != gotPass {
		http.Error(w, "invalid password", http.StatusForbidden)
		return
	}
	h.h.ServeHTTP(w, r)
}

// plan9LogWriter truncates log writes to 128 bytes,
// to work around some Plan 9 and/or GCE serial port bug.
type plan9LogWriter struct {
	w   io.Writer
	buf []byte
}

func (pw *plan9LogWriter) Write(p []byte) (n int, err error) {
	const max = 128 - len("\n\x00")
	if len(p) < max {
		return pw.w.Write(p)
	}
	if pw.buf == nil {
		pw.buf = make([]byte, max+1)
	}
	n = copy(pw.buf[:max], p)
	pw.buf[n] = '\n'
	return pw.w.Write(pw.buf[:n+1])
}

func requireTrailerSupport() {
	// Depend on a symbol that was added after HTTP Trailer support was
	// implemented (4b96409 Dec 29 2014)j so that this function will fail
	// to compile without Trailer support.
	// bufio.Reader.Discard was added by ee2ecc4 Jan 7 2015.
	var r bufio.Reader
	_ = r.Discard
}

var killProcessTree = killProcessTreeUnix

func killProcessTreeUnix(p *os.Process) error {
	return p.Kill()
}
