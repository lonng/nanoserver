// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// racebuild builds the race runtime (syso files) on all supported OSes using gomote.
// Usage:
//	$ racebuild -rev <llvm_revision> -goroot <path_to_go_repo>
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	flagGoroot = flag.String("goroot", "", "path to Go repository to update (required)")
	flagRev    = flag.String("rev", "", "llvm compiler-rt git revision from http://llvm.org/git/compiler-rt.git (required)")
)

// TODO: use buildlet package instead of calling out to gomote.
var platforms = []*Platform{
	&Platform{
		OS:   "freebsd",
		Arch: "amd64",
		Type: "freebsd-amd64-race",
		Script: `#!/usr/bin/env bash
set -e
git clone https://go.googlesource.com/go
git clone http://llvm.org/git/compiler-rt.git
(cd compiler-rt && git checkout $REV)
(cd compiler-rt/lib/tsan/go && CC=clang ./buildgo.sh)
cp compiler-rt/lib/tsan/go/race_freebsd_amd64.syso go/src/runtime/race
(cd go/src && ./race.bash)
			`,
	},
	&Platform{
		OS:   "darwin",
		Arch: "amd64",
		Type: "darwin-amd64-10_10",
		Script: `#!/usr/bin/env bash
set -e
git clone https://go.googlesource.com/go
git clone http://llvm.org/git/compiler-rt.git
(cd compiler-rt && git checkout $REV)
(cd compiler-rt/lib/tsan/go && CC=clang ./buildgo.sh)
cp compiler-rt/lib/tsan/go/race_darwin_amd64.syso go/src/runtime/race
(cd go/src && ./race.bash)
			`,
	},
	&Platform{
		OS:   "linux",
		Arch: "amd64",
		Type: "linux-amd64-race",
		Script: `#!/usr/bin/env bash
set -e
apt-get update
apt-get install -y git g++
git clone https://go.googlesource.com/go
git clone http://llvm.org/git/compiler-rt.git
(cd compiler-rt && git checkout $REV)
(cd compiler-rt/lib/tsan/go && ./buildgo.sh)
cp compiler-rt/lib/tsan/go/race_linux_amd64.syso go/src/runtime/race
(cd go/src && ./race.bash)
			`,
	},
	{
		OS:   "windows",
		Arch: "amd64",
		Type: "windows-amd64-race",
		Script: `
	git clone https://go.googlesource.com/go
	if %errorlevel% neq 0 exit /b %errorlevel%
	git clone http://llvm.org/git/compiler-rt.git
	if %errorlevel% neq 0 exit /b %errorlevel%
	cd compiler-rt
	git checkout %REV%
	if %errorlevel% neq 0 exit /b %errorlevel%
	cd ..
	cd compiler-rt/lib/tsan/go
	call build.bat
	if %errorlevel% neq 0 exit /b %errorlevel%
	cd ../../../..
	xcopy compiler-rt\lib\tsan\go\race_windows_amd64.syso go\src\runtime\race\race_windows_amd64.syso /Y
	if %errorlevel% neq 0 exit /b %errorlevel%
	cd go/src
	call race.bat
	if %errorlevel% neq 0 exit /b %errorlevel%
				`,
	},
}

func main() {
	flag.Parse()
	if *flagRev == "" || *flagGoroot == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Update revision in the README file.
	// Do this early to check goroot correctness.
	readmeFile := filepath.Join(*flagGoroot, "src", "runtime", "race", "README")
	readme, err := ioutil.ReadFile(readmeFile)
	if err != nil {
		log.Fatalf("bad -goroot? %v", err)
	}
	readmeRev := regexp.MustCompile("Current runtime is built on rev ([0-9,a-z]+)\\.").FindSubmatchIndex(readme)
	if readmeRev == nil {
		log.Fatalf("failed to find current revision in src/runtime/race/README")
	}
	readme = bytes.Replace(readme, readme[readmeRev[2]:readmeRev[3]], []byte(*flagRev), -1)
	if err := ioutil.WriteFile(readmeFile, readme, 0640); err != nil {
		log.Fatalf("failed to write README file: %v", err)
	}

	// Start build on all platforms in parallel.
	var wg sync.WaitGroup
	wg.Add(len(platforms))
	for _, p := range platforms {
		p := p
		go func() {
			defer wg.Done()
			p.Err = p.Build()
			if p.Err != nil {
				p.Err = fmt.Errorf("failed: %v", p.Err)
				log.Printf("%v: %v", p.Name, p.Err)
			}
		}()
	}
	wg.Wait()

	// Duplicate results, they can get lost in the log.
	ok := true
	log.Printf("---")
	for _, p := range platforms {
		if p.Err == nil {
			log.Printf("%v: ok", p.Name)
			continue
		}
		ok = false
		log.Printf("%v: %v", p.Name, p.Err)
	}
	if !ok {
		os.Exit(1)
	}
}

type Platform struct {
	OS     string
	Arch   string
	Name   string // something for logging
	Type   string // gomote instance type
	Inst   string // actual gomote instance name
	Err    error
	Log    *os.File
	Script string
}

func (p *Platform) Build() error {
	p.Name = fmt.Sprintf("%v-%v", p.OS, p.Arch)

	// Open log file.
	var err error
	p.Log, err = ioutil.TempFile("", p.Name)
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}
	defer p.Log.Close()
	log.Printf("%v: logging to %v", p.Name, p.Log.Name())

	// Create gomote instance (or reuse an existing instance for debugging).
	if p.Inst == "" {
		// Creation sometimes fails with transient errors like:
		// "buildlet didn't come up at http://10.240.0.13 in 3m0s".
		var createErr error
		for i := 0; i < 10; i++ {
			inst, err := p.Gomote("create", p.Type)
			if err != nil {
				log.Printf("%v: instance creation failed, retrying", p.Name)
				createErr = err
				continue
			}
			p.Inst = strings.Trim(string(inst), " \t\n")
			break
		}
		if p.Inst == "" {
			return createErr
		}
	}
	log.Printf("%s: using instance %v", p.Name, p.Inst)

	// put14
	if _, err := p.Gomote("put14", p.Inst); err != nil {
		return err
	}

	// Execute the script.
	script, err := ioutil.TempFile("", "racebuild")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer func() {
		script.Close()
		os.Remove(script.Name())
	}()
	if _, err := script.Write([]byte(p.Script)); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}
	script.Close()
	targetName := "script.bash"
	if p.OS == "windows" {
		targetName = "script.bat"
	}
	if _, err := p.Gomote("put", "-mode=0700", p.Inst, script.Name(), targetName); err != nil {
		return err
	}
	if _, err := p.Gomote("run", "-e=REV="+*flagRev, p.Inst, targetName); err != nil {
		return err
	}

	// The script is supposed to leave updated runtime at that path. Copy it out.
	syso := fmt.Sprintf("race_%v_%s.syso", p.OS, p.Arch)
	targz, err := p.Gomote("gettar", "-dir=go/src/runtime/race/"+syso, p.Inst)
	if err != nil {
		return err
	}

	// Untar the runtime and write it to goroot.
	if err := p.WriteSyso(filepath.Join(*flagGoroot, "src", "runtime", "race", syso), targz); err != nil {
		return fmt.Errorf("%v", err)
	}

	log.Printf("%v: build completed", p.Name)
	return nil
}

func (p *Platform) WriteSyso(sysof string, targz []byte) error {
	// Ungzip.
	gzipr, err := gzip.NewReader(bytes.NewReader(targz))
	if err != nil {
		return fmt.Errorf("failed to read gzip archive: %v", err)
	}
	defer gzipr.Close()
	tr := tar.NewReader(gzipr)
	if _, err := tr.Next(); err != nil {
		return fmt.Errorf("failed to read tar archive: %v", err)
	}

	// Copy the file.
	syso, err := os.Create(sysof)
	if err != nil {
		return fmt.Errorf("failed to open race runtime: %v", err)
	}
	defer syso.Close()
	if _, err := io.Copy(syso, tr); err != nil {
		return fmt.Errorf("failed to write race runtime: %v", err)
	}
	return nil
}

func (p *Platform) Gomote(args ...string) ([]byte, error) {
	log.Printf("%v: gomote %v", p.Name, args)
	fmt.Fprintf(p.Log, "$ gomote %v\n", args)
	output, err := exec.Command("gomote", args...).CombinedOutput()
	if err != nil || args[0] != "gettar" {
		p.Log.Write(output)
	}
	fmt.Fprintf(p.Log, "\n\n")
	if err != nil {
		err = fmt.Errorf("gomote %v failed: %v", args, err)
	}
	return output, err
}
