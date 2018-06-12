// +build sqlite3

// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	supportedDrivers["sqlite3"] = "github.com/mattn/go-sqlite3"
}
