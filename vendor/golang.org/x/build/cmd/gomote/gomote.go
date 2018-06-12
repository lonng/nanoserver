// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
The gomote command is a client for the Go builder infrastructure.
It's a remote control for remote Go builder machines.

Usage:

  gomote [global-flags] cmd [cmd-flags]

  For example,
  $ gomote create openbsd-amd64-gce56
  $ gomote push
  $ gomote run openbsd-amd64-gce56 src/make.bash

TODO: document more, and figure out the CLI interface more.
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"golang.org/x/build"
	"golang.org/x/build/buildenv"
	"golang.org/x/build/buildlet"
)

var (
	user = flag.String("user", username(), "gomote server username. You must have the token for user-$USER. The error message will say where to put it.")

	staging = flag.Bool("staging", false, "if true, use the staging build coordinator and buildlets")

	buildEnv = buildenv.Production
)

type command struct {
	name string
	des  string
	run  func([]string) error
}

var commands = map[string]command{}

func sortedCommands() []string {
	s := make([]string, 0, len(commands))
	for name := range commands {
		s = append(s, name)
	}
	sort.Strings(s)
	return s
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage of gomote: gomote [global-flags] <cmd> [cmd-flags]

Global flags:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "Commands:\n\n")
	for _, name := range sortedCommands() {
		fmt.Fprintf(os.Stderr, "  %-10s %s\n", name, commands[name].des)
	}
	os.Exit(1)
}

func registerCommand(name, des string, run func([]string) error) {
	if _, dup := commands[name]; dup {
		panic("duplicate registration of " + name)
	}
	commands[name] = command{
		name: name,
		des:  des,
		run:  run,
	}
}

func coordinatorClient() *buildlet.CoordinatorClient {
	inst := build.ProdCoordinator
	if *staging {
		inst = build.StagingCoordinator
	}
	return &buildlet.CoordinatorClient{
		Auth: buildlet.UserPass{
			Username: "user-" + *user,
			Password: userToken(),
		},
		Instance: inst,
	}
}

func registerCommands() {
	registerCommand("create", "create a buildlet", create)
	registerCommand("destroy", "destroy a buildlet", destroy)
	registerCommand("gettar", "extract a tar.gz from a buildlet", getTar)
	registerCommand("ls", "list the contents of a directory on a buildlet", ls)
	registerCommand("list", "list buildlets", list)
	registerCommand("ping", "test whether a buildlet is alive and reachable ", ping)
	registerCommand("push", "sync the repo of your pwd to the buildlet", push)
	registerCommand("put", "put files on a buildlet", put)
	registerCommand("put14", "put Go 1.4 in place", put14)
	registerCommand("puttar", "extract a tar.gz to a buildlet", putTar)
	registerCommand("rm", "delete files or directories", rm)
	registerCommand("run", "run a command on a buildlet", run)
}

func main() {
	registerCommands()
	flag.Usage = usage
	flag.Parse()
	if *staging {
		buildEnv = buildenv.Staging
	}
	args := flag.Args()
	if len(args) == 0 {
		usage()
	}
	cmdName := args[0]
	cmd, ok := commands[cmdName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmdName)
		usage()
	}
	err := cmd.run(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running %s: %v\n", cmdName, err)
		os.Exit(1)
	}
}
