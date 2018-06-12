// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package godash

import (
	"net/http"
	"sort"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const project = "golang/go"
const projectOwner = "golang"
const projectRepo = "go"

func NewGitHubClient(project, authToken string, transport http.RoundTripper) *github.Client {
	t := &oauth2.Transport{
		Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: authToken}),
		Base:   transport,
	}
	return github.NewClient(&http.Client{Transport: t})
}

func getInt(x *int) int {
	if x == nil {
		return 0
	}
	return *x
}

func getString(x *string) string {
	if x == nil {
		return ""
	}
	return *x
}

func getUserLogin(x *github.User) string {
	if x == nil || x.Login == nil {
		return ""
	}
	return *x.Login
}

func getTime(x *time.Time) time.Time {
	if x == nil {
		return time.Time{}
	}
	return (*x).Local()
}

func getMilestoneTitle(x *github.Milestone) string {
	if x == nil || x.Title == nil {
		return ""
	}
	return *x.Title
}

func getLabelNames(x []github.Label) []string {
	var out []string
	for _, lab := range x {
		out = append(out, getString(lab.Name))
	}
	sort.Strings(out)
	return out
}

func issueToIssue(issue github.Issue) *Issue {
	return &Issue{
		Number:    getInt(issue.Number),
		Title:     getString(issue.Title),
		State:     getString(issue.State),
		Assignee:  getUserLogin(issue.Assignee),
		Closed:    getTime(issue.ClosedAt),
		Labels:    getLabelNames(issue.Labels),
		Milestone: getMilestoneTitle(issue.Milestone),
		Reporter:  getUserLogin(issue.User),
		Created:   getTime(issue.CreatedAt),
	}
}

func listIssues(client *github.Client, opt github.IssueListByRepoOptions) ([]*Issue, error) {
	var all []*Issue
	for page := 1; ; {
		xopt := opt
		xopt.ListOptions = github.ListOptions{
			Page:    page,
			PerPage: 100,
		}
		issues, resp, err := client.Issues.ListByRepo(projectOwner, projectRepo, &xopt)
		for _, issue := range issues {
			if issue.PullRequestLinks == nil {
				all = append(all, issueToIssue(issue))
			}
		}
		if err != nil {
			return all, err
		}
		if resp.NextPage < page {
			break
		}
		page = resp.NextPage
	}
	return all, nil
}

func searchIssues(client *github.Client, q string) ([]*Issue, error) {
	var all []*Issue
	for page := 1; ; {
		// TODO(rsc): Rethink excluding pull requests.
		x, resp, err := client.Search.Issues("type:issue state:open repo:"+project+" "+q, &github.SearchOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
		for _, issue := range x.Issues {
			all = append(all, issueToIssue(issue))
		}
		if err != nil {
			return all, err
		}
		if resp.NextPage < page {
			break
		}
		page = resp.NextPage
	}
	return all, nil
}

func getMilestones(client *github.Client) ([]*github.Milestone, error) {
	var all []*github.Milestone
	milestones, _, err := client.Issues.ListMilestones(projectOwner, projectRepo, nil)
	for i := range milestones {
		m := &milestones[i]
		if m.Title != nil {
			all = append(all, m)
		}
	}
	return all, err
}
