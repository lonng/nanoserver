// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/build/buildlet"
	"golang.org/x/build/dashboard"
	"golang.org/x/build/envutil"
)

func run(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "create usage: gomote run [run-opts] <instance> <cmd> [args...]")
		fs.PrintDefaults()
		os.Exit(1)
	}
	var sys bool
	fs.BoolVar(&sys, "system", false, "run inside the system, and not inside the workdir; this is implicit if cmd starts with '/'")
	var debug bool
	fs.BoolVar(&debug, "debug", false, "write debug info about the command's execution before it begins")
	var env stringSlice
	fs.Var(&env, "e", "Environment variable KEY=value. The -e flag may be repeated multiple times to add multiple things to the environment.")
	var path string
	fs.StringVar(&path, "path", "", "Comma-separated list of ExecOpts.Path elements. The special string 'EMPTY' means to run without any $PATH. The empty string (default) does not modify the $PATH.")
	var dir string
	fs.StringVar(&dir, "dir", "", "Directory to run from. Defaults to the directory of the command, or the work directory if -system is true.")

	fs.Parse(args)
	if fs.NArg() < 2 {
		fs.Usage()
	}
	name, cmd := fs.Arg(0), fs.Arg(1)

	var conf dashboard.BuildConfig

	bc, conf, err := clientAndConf(name)
	if err != nil {
		return err
	}

	var pathOpt []string
	if path == "EMPTY" {
		pathOpt = []string{} // non-nil
	} else if path != "" {
		pathOpt = strings.Split(path, ",")
	}

	remoteErr, execErr := bc.Exec(cmd, buildlet.ExecOpts{
		Dir:         dir,
		SystemLevel: sys || strings.HasPrefix(cmd, "/"),
		Output:      os.Stdout,
		Args:        fs.Args()[2:],
		ExtraEnv:    envutil.Dedup(conf.GOOS() == "windows", append(conf.Env(), []string(env)...)),
		Debug:       debug,
		Path:        pathOpt,
	})
	if execErr != nil {
		return fmt.Errorf("Error trying to execute %s: %v", cmd, execErr)
	}
	return remoteErr
}

// stringSlice implements flag.Value, specifically for storing environment
// variable key=value pairs.
type stringSlice []string

func (*stringSlice) String() string { return "" } // default value

func (ss *stringSlice) Set(v string) error {
	if v != "" {
		if !strings.Contains(v, "=") {
			return fmt.Errorf("-e argument %q doesn't contains an '=' sign.", v)
		}
		*ss = append(*ss, v)
	}
	return nil
}
