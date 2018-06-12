// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/build/buildlet"
)

func push(args []string) error {
	fs := flag.NewFlagSet("push", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "create usage: gomote push <instance>")
		fs.PrintDefaults()
		os.Exit(1)
	}
	fs.Parse(args)

	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		slurp, err := exec.Command("go", "env", "GOROOT").Output()
		if err != nil {
			return fmt.Errorf("failed to get GOROOT from go env: %v", err)
		}
		goroot = strings.TrimSpace(string(slurp))
		if goroot == "" {
			return errors.New("Failed to get $GOROOT from environment or go env")
		}
	}

	if fs.NArg() != 1 {
		fs.Usage()
	}
	name := fs.Arg(0)
	bc, conf, err := clientAndConf(name)
	if err != nil {
		return err
	}

	haveGo14 := false
	haveGo := false
	remote := map[string]buildlet.DirEntry{} // keys like "src/make.bash"

	lsOpts := buildlet.ListDirOpts{
		Recursive: true,
		Digest:    true,
		Skip: []string{
			// Ignore binary output directories:
			"go/pkg", "go/bin",
			// We don't care about the digest of
			// particular source files for Go 1.4.  And
			// exclude /pkg. This leaves go1.4/bin, which
			// is enough to know whether we have Go 1.4 or
			// not.
			"go1.4/src", "go1.4/pkg",
		},
	}
	if err := bc.ListDir(".", lsOpts, func(ent buildlet.DirEntry) {
		name := ent.Name()
		if strings.HasPrefix(name, "go1.4/") {
			haveGo14 = true
			return
		}
		if strings.HasPrefix(name, "go/") {
			haveGo = true
			remote[name[len("go/"):]] = ent
		}

	}); err != nil {
		return fmt.Errorf("error listing buildlet's existing files: %v", err)
	}

	if !haveGo14 {
		if u := conf.GoBootstrapURL(buildEnv); u != "" {
			log.Printf("installing go1.4")
			if err := bc.PutTarFromURL(u, "go1.4"); err != nil {
				return err
			}
		}
	}

	// TODO(bradfitz,adg): if !haveGo, then run 'git rev-parse
	// HEAD' in goroot and see what the user's HEAD is,
	// approximately. Then tell the buildlet to fetch that tarball
	// directly from Gerrit as a base. Then do another ListDir to
	// the buildlet to get the new list, which will result in a
	// smaller upload payload. This is important when working from
	// home with a slower network connection to GCE. (upload a few
	// KB instead of 10-40 MB)

	type fileInfo struct {
		fi   os.FileInfo
		sha1 string // if regular file
	}
	local := map[string]fileInfo{} // keys like "src/make.bash"
	if err := filepath.Walk(goroot, func(path string, fi os.FileInfo, err error) error {
		if isEditorBackup(path) {
			return nil
		}
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(filepath.ToSlash(strings.TrimPrefix(path, goroot)), "/")
		if rel == "" {
			return nil
		}
		if strings.HasPrefix(rel, "test/") && isDotArchChar(rel) {
			// Skip test/bench/shootout/spectral-norm.6, etc.
			return nil
		}
		if fi.IsDir() {
			switch rel {
			case ".git", "pkg", "bin":
				return filepath.SkipDir
			}
		}
		inf := fileInfo{fi: fi}
		if fi.Mode().IsRegular() {
			inf.sha1, err = fileSHA1(path)
			if err != nil {
				return err
			}
		}
		local[rel] = inf
		return nil
	}); err != nil {
		return fmt.Errorf("error enumerating local GOROOT files: %v", err)
	}

	var toDel []string
	for rel := range remote {
		if rel == "VERSION" {
			// Don't delete this. It's harmless, and necessary.
			// Clients can overwrite it if they want.
			continue
		}
		rel = strings.TrimRight(rel, "/")
		if rel == "" {
			continue
		}
		if _, ok := local[rel]; !ok {
			toDel = append(toDel, rel)
		}
	}
	if len(toDel) > 0 {
		withGo := make([]string, len(toDel)) // with the "go/" prefix
		for i, v := range toDel {
			withGo[i] = "go/" + v
		}
		sort.Strings(withGo)
		log.Printf("Deleting remote files: %q", withGo)
		if err := bc.RemoveAll(withGo...); err != nil {
			return fmt.Errorf("Deleting remote unwanted files: %v", err)
		}
	}

	var toSend []string
	for rel, inf := range local {
		if !inf.fi.Mode().IsRegular() {
			// TODO(bradfitz): this is only doing regular files
			// for now, so empty directories, symlinks, etc aren't
			// supported. revisit if that's a problem.
			if !inf.fi.IsDir() {
				log.Printf("Ignoring local non-regular, non-directory file %s: %v", rel, inf.fi.Mode())
			}
			continue
		}
		rem, ok := remote[rel]
		if !ok {
			log.Printf("Remote doesn't have %q", rel)
			toSend = append(toSend, rel)
			continue
		}
		if rem.Digest() != inf.sha1 {
			log.Printf("Remote's %s digest is %q; want %q", rel, rem.Digest(), inf.sha1)
			toSend = append(toSend, rel)
		}
	}
	if _, hasVersion := remote["VERSION"]; !hasVersion {
		log.Printf("Remote lacks a VERSION file; sending a fake one")
		toSend = append(toSend, "VERSION")
	}
	if len(toSend) > 0 {
		sort.Strings(toSend)
		tgz, err := generateDeltaTgz(goroot, toSend)
		if err != nil {
			return err
		}
		log.Printf("Uploading %d new/changed files; %d byte .tar.gz", len(toSend), tgz.Len())
		if err := bc.PutTar(tgz, "go"); err != nil {
			return fmt.Errorf("writing tarball to buildlet: %v", err)
		}
	}
	return nil
}

func isEditorBackup(path string) bool {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") && strings.HasSuffix(base, ".swp") {
		// vi
		return true
	}
	if strings.HasSuffix(path, "~") || strings.HasSuffix(path, "#") ||
		strings.HasPrefix(base, "#") || strings.HasPrefix(base, ".#") {
		// emacs
		return true
	}
	return false
}

// file is forward-slash separated
func generateDeltaTgz(goroot string, files []string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(zw)
	for _, file := range files {
		// Special.
		if file == "VERSION" {
			// TODO(bradfitz): a dummy VERSION file's contents to make things
			// happy. Notably it starts with "devel ". Do we care about it
			// being accurate beyond that?
			version := "devel gomote.XXXXX"
			if err := tw.WriteHeader(&tar.Header{
				Name: "VERSION",
				Mode: 0644,
				Size: int64(len(version)),
			}); err != nil {
				return nil, err
			}
			if _, err := io.WriteString(tw, version); err != nil {
				return nil, err
			}
			continue
		}
		f, err := os.Open(filepath.Join(goroot, file))
		if err != nil {
			return nil, err
		}
		fi, err := f.Stat()
		if err != nil {
			f.Close()
			return nil, err
		}
		header, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			f.Close()
			return nil, err
		}
		header.Name = file // forward slash
		if err := tw.WriteHeader(header); err != nil {
			f.Close()
			return nil, err
		}
		if _, err := io.CopyN(tw, f, header.Size); err != nil {
			f.Close()
			return nil, fmt.Errorf("error copying contents of %s: %v", file, err)
		}
		f.Close()
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
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

func isDotArchChar(path string) bool {
	if len(path) < 2 || path[len(path)-2] != '.' {
		return false
	}
	switch path[len(path)-1] {
	case '6', '8', '5', '7', '9':
		return true
	}
	return false
}
