// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/build/buildlet"
	"golang.org/x/build/dashboard"
)

func list(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "list usage: gomote list")
		fs.PrintDefaults()
		os.Exit(1)
	}
	fs.Parse(args)
	if fs.NArg() != 0 {
		fs.Usage()
	}

	cc := coordinatorClient()
	rbs, err := cc.RemoteBuildlets()
	if err != nil {
		log.Fatal(err)
	}
	for _, rb := range rbs {
		fmt.Printf("%s\t%s\texpires in %v\n", rb.Name, rb.Type, rb.Expires.Sub(time.Now()))
	}

	return nil
}

func clientAndConf(name string) (bc *buildlet.Client, conf dashboard.BuildConfig, err error) {
	cc := coordinatorClient()

	rbs, err := cc.RemoteBuildlets()
	if err != nil {
		return
	}
	var ok bool
	for _, rb := range rbs {
		if rb.Name == name {
			conf, ok = namedConfig(rb.Type)
			if !ok {
				err = fmt.Errorf("builder %q exists, but unknown type %q", name, rb.Type)
				return
			}
			break
		}
	}
	if !ok {
		err = fmt.Errorf("unknown builder %q", name)
		return
	}

	bc, err = namedClient(name)
	return
}

func namedClient(name string) (*buildlet.Client, error) {
	if strings.Contains(name, ":") {
		return buildlet.NewClient(name, buildlet.NoKeyPair), nil
	}
	cc := coordinatorClient()
	return cc.NamedBuildlet(name)
}

// namedConfig returns the builder configuration that matches the given mote
// name. It matches prefixes to accommodate motes than have "-n" suffixes.
func namedConfig(name string) (dashboard.BuildConfig, bool) {
	match := ""
	for cname := range dashboard.Builders {
		if strings.HasPrefix(name, cname) && len(cname) > len(match) {
			match = cname
		}
	}
	return dashboard.Builders[match], match != ""
}
