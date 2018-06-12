// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package godash

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Adapted from git-codereview/mail.go, but uses Author lines
// in addition to Reviewed-By lines. The effect should be the same,
// since the most common reviewers are the most common authors too,
// but admitting authors lets us shorten CL owners too.

func (r *Reviewers) LoadLocal() {
	output, err := exec.Command("go", "env", "GOROOT").CombinedOutput()
	if err != nil {
		log.Fatalf("go env GOROOT: %v\n%s", err, output)
	}
	goroot := strings.TrimSpace(string(output))
	// TODO(quentin): This should probably look at origin/master, not master.
	cmd := exec.Command("git", "log", "--format=format:Author: <%aE>%n%B")
	cmd.Dir = goroot
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("git log: %v\n%s", err, output)
	}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "Reviewed-by:") || strings.HasPrefix(line, "Author:") {
			f := strings.Fields(line)
			addr := f[len(f)-1]
			if strings.HasPrefix(addr, "<") && strings.Contains(addr, "@") && strings.HasSuffix(addr, ">") {
				email := addr[1 : len(addr)-1]
				r.add(email, strings.HasPrefix(line, "Reviewed-by:"))
			}
		}
	}
	cmd = exec.Command("git", "show", "--format=%H%n%ct", "--no-patch")
	cmd.Dir = goroot
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("git show: %v\n%s", err, output)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) != 2 {
		log.Fatalf("git show: failed to parse\n%s", output)
	}
	r.data.LastSHA = lines[0]
	t, err := strconv.ParseInt(lines[1], 10, 64)
	if err != nil {
		log.Fatalf("git show: failed to parse time %q: %v", lines[1], err)
	}
	r.data.LastTime = time.Unix(t, 0)
	r.recalculate()
}
