// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package godash

import (
	"strings"

	"github.com/google/go-github/github"
)

// LoadGithub fetches the list of reviewers for the current
// repository, sorted by how many reviews each has done.
func (r *Reviewers) LoadGithub(client *github.Client) error {
	// TODO(quentin): GitHub has an API for fetching the list of
	// authors ordered by # of commits... is this sufficient
	// instead of walking the whole commit list?
	lastTime, lastSHA := r.data.LastTime, r.data.LastSHA
pages:
	for page := 1; ; {
		opt := github.CommitsListOptions{
			SHA: "master",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		}
		commits, resp, err := client.Repositories.ListCommits(projectOwner, projectRepo, &opt)
		for _, commit := range commits {
			if commit.Commit != nil && commit.Commit.Committer != nil && commit.Commit.Committer.Date != nil && commit.SHA != nil {
				if commit.Commit.Committer.Date.After(r.data.LastTime) {
					r.data.LastTime = *commit.Commit.Committer.Date
					r.data.LastSHA = *commit.SHA
				}
				if commit.Commit.Committer.Date.Before(lastTime) || *commit.SHA == lastSHA {
					break pages
				}
			}
			if commit.Commit != nil && commit.Commit.Author != nil && commit.Commit.Author.Email != nil {
				r.add(*commit.Commit.Author.Email, false)
			}
			if commit.Commit != nil && commit.Commit.Message != nil {
				for _, line := range strings.Split(string(*commit.Commit.Message), "\n") {
					if strings.HasPrefix(line, "Reviewed-by:") {
						f := strings.Fields(line)
						addr := f[len(f)-1]
						if strings.HasPrefix(addr, "<") && strings.HasSuffix(addr, ">") {
							addr = addr[1 : len(addr)-1]
						}
						r.add(addr, true)
					}
				}
			}
		}
		if err != nil {
			return err
		}
		if resp.NextPage < page {
			break
		}
		page = resp.NextPage
	}
	r.recalculate()
	return nil
}
