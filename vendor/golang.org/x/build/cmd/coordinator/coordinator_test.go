// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"strings"
	"testing"
)

func TestPartitionGoTests(t *testing.T) {
	var in []string
	for name := range fixedTestDuration {
		in = append(in, name)
	}
	sets := partitionGoTests(in)
	for i, set := range sets {
		t.Logf("set %d = \"-run=^(%s)$\"", i, strings.Join(set, "|"))
	}
}
