// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package godash

import "golang.org/x/build/gerrit"

func fetchCLs(client *gerrit.Client, reviewers *Reviewers, goReleaseCycle int, q string) ([]*CL, error) {
	cis, err := client.QueryChanges("-project:scratch -message:do-not-review "+q, gerrit.QueryChangesOpt{
		N: 5000,
		Fields: []string{
			"LABELS",
			"CURRENT_FILES",
			"CURRENT_REVISION",
			"CURRENT_COMMIT",
			"MESSAGES",
			"DETAILED_ACCOUNTS", // fill out Owner.AuthorInfo, etc
			"DETAILED_LABELS",
		},
	})
	if err != nil {
		return nil, err
	}
	cls := []*CL{}
	for _, ci := range cis {
		cl := ParseCL(ci, reviewers, goReleaseCycle)
		if cl.Closed || cl.DoNotReview {
			continue
		}
		cls = append(cls, cl)
	}
	return cls, nil
}
