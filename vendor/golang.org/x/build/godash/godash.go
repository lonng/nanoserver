// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package godash generates dashboards about issues and CLs in the Go
// Github and Gerrit projects. There is a user-friendly interface in
// the godash command-line tool at golang.org/x/build/cmd/godash
package godash

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/build/gerrit"
)

type Issue struct {
	Number          int
	Title           string
	Labels          []string
	Reporter        string
	Assignee        string
	Milestone       string
	State           string
	Created, Closed time.Time
}

type Group struct {
	Dir   string
	Items []*Item
}

type Item struct {
	Issue *Issue
	CLs   []*CL
}

// Data represents all the data needed to compute the dashboard
type Data struct {
	Issues         map[int]*Issue
	CLs            []*CL
	Milestones     []*github.Milestone
	GoReleaseCycle int
	Now            time.Time

	Reviewers *Reviewers
}

var (
	releaseRE = regexp.MustCompile(`^Go1\.(\d+)$`)
)

func (d *Data) FetchData(gh *github.Client, ger *gerrit.Client, days int, clOnly, includeMerged bool) error {
	d.Now = time.Now()
	m, err := getMilestones(gh)
	if err != nil {
		return err
	}
	d.Milestones = m

	for _, m := range d.Milestones {
		if matches := releaseRE.FindStringSubmatch(*m.Title); matches != nil {
			n, _ := strconv.Atoi(matches[1])
			if d.GoReleaseCycle == 0 || d.GoReleaseCycle > n {
				d.GoReleaseCycle = n
			}
		}
	}
	since := d.Now.Add(-(time.Duration(days)*24 + 12) * time.Hour).UTC().Round(time.Second)
	cls, err := fetchCLs(ger, d.Reviewers, d.GoReleaseCycle, "is:open")
	if err != nil {
		return err
	}

	var open []*CL
	for _, cl := range cls {
		if !cl.Closed && (clOnly || !strings.HasPrefix(cl.Subject, "[dev.")) {
			open = append(open, cl)
		}
	}
	if includeMerged {
		cls, err := fetchCLs(ger, d.Reviewers, d.GoReleaseCycle, "is:merged since:\""+since.Format("2006-01-02 15:04:05")+"\"")
		if err != nil {
			return err
		}
		open = append(open, cls...)
	}
	d.CLs = open

	d.Issues = make(map[int]*Issue)

	if !clOnly {
		res, err := listIssues(gh, github.IssueListByRepoOptions{State: "open"})
		if err != nil {
			return err
		}
		res2, err := searchIssues(gh, "is:closed closed:>="+since.Format(time.RFC3339))
		if err != nil {
			return err
		}
		res = append(res, res2...)
		for _, issue := range res {
			d.Issues[issue.Number] = issue
		}
	}
	return nil
}

func (d *Data) groupData(includeIssues, allCLs bool) []*Group {
	groupsByDir := make(map[string]*Group)
	addGroup := func(item *Item) {
		dir := item.Dir()
		g := groupsByDir[dirKey(dir)]
		if g == nil {
			g = &Group{Dir: dir}
			groupsByDir[dirKey(dir)] = g
		}
		g.Items = append(g.Items, item)
	}
	itemsByBug := map[int]*Item{}

	if includeIssues {
		for _, issue := range d.Issues {
			item := &Item{Issue: issue}
			addGroup(item)
			itemsByBug[issue.Number] = item
		}
	}

	for _, cl := range d.CLs {
		found := false
		for _, id := range cl.Issues {
			item := itemsByBug[id]
			if item != nil {
				found = true
				item.CLs = append(item.CLs, cl)
			}
		}
		if !found {
			if cl.Project == "go" || allCLs {
				item := &Item{CLs: []*CL{cl}}
				addGroup(item)
			}
		}
	}

	var keys []string
	for key, g := range groupsByDir {
		sort.Sort(itemsBySummary(g.Items))
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var groups []*Group
	for _, key := range keys {
		g := groupsByDir[key]
		groups = append(groups, g)
	}
	return groups
}

var okDesc = map[string]bool{
	"all":   true,
	"build": true,
}

func (item *Item) Dir() string {
	for _, cl := range item.CLs {
		if cl.Status == "merged" {
			return "closed"
		}
		dirs := cl.Dirs()
		desc := titleDir(cl.Subject)

		// Accept description if it is a global prefix like "all".
		if okDesc[desc] {
			return desc
		}

		// Accept description if it matches one of the directories.
		for _, dir := range dirs {
			if dir == desc {
				return dir
			}
		}

		// Otherwise use most common directory.
		if len(dirs) > 0 {
			return dirs[0]
		}

		// Otherwise accept description.
		return desc
	}
	if item.Issue != nil {
		if item.Issue.State == "closed" {
			return "closed"
		}
		if hasLabel(item.Issue, "Proposal") {
			return "proposal"
		}
		if dir := titleDir(item.Issue.Title); dir != "" {
			return dir
		}
		return "?"
	}
	return "?"
}

func hasLabel(issue *Issue, label string) bool {
	for _, lab := range issue.Labels {
		if label == lab {
			return true
		}
	}
	return false
}

func titleDir(title string) string {
	if i := strings.Index(title, "\n"); i >= 0 {
		title = title[:i]
	}
	title = strings.TrimSpace(title)
	i := strings.Index(title, ":")
	if i < 0 {
		return ""
	}
	title = title[:i]
	if i := strings.Index(title, ","); i >= 0 {
		title = strings.TrimSpace(title[:i])
	}
	if strings.Contains(title, " ") {
		return ""
	}
	return title
}

// Dirs returns the list of directories that this CL might be said to be about,
// in preference order.
func (cl *CL) Dirs() []string {
	prefix := ""
	if cl.Project != "go" {
		prefix = "x/" + cl.Project + "/"
	}
	counts := map[string]int{}
	for _, file := range cl.Files {
		name := file
		i := strings.LastIndex(name, "/")
		if i >= 0 {
			name = name[:i]
		} else {
			name = ""
		}
		name = strings.TrimPrefix(name, "src/")
		if name == "src" {
			name = ""
		}
		name = prefix + name
		if name == "" {
			name = "build"
		}
		counts[name]++
	}

	if _, ok := counts["test"]; ok {
		counts["test"] -= 10000 // do not pick as most frequent
	}

	var dirs dirCounts
	for name, count := range counts {
		dirs = append(dirs, dirCount{name, count})
	}
	sort.Sort(dirs)

	var names []string
	for _, d := range dirs {
		names = append(names, d.name)
	}
	return names
}

type dirCount struct {
	name  string
	count int
}

type dirCounts []dirCount

func (x dirCounts) Len() int      { return len(x) }
func (x dirCounts) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
func (x dirCounts) Less(i, j int) bool {
	if x[i].count != x[j].count {
		return x[i].count > x[j].count
	}
	return x[i].name < x[j].name
}

type itemsBySummary []*Item

func (x itemsBySummary) Len() int           { return len(x) }
func (x itemsBySummary) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x itemsBySummary) Less(i, j int) bool { return itemSummary(x[i]) < itemSummary(x[j]) }

func itemSummary(it *Item) string {
	if it.Issue != nil {
		return it.Issue.Title
	}
	for _, cl := range it.CLs {
		return cl.Subject
	}
	return ""
}

func dirKey(s string) string {
	if strings.Contains(s, ".") {
		return "\x7F" + s
	}
	return s
}

var milestoneRE = regexp.MustCompile(`^Go1\.(\d+)(|\.(\d+))(|[A-Z].*)$`)

type milestone struct {
	title        string
	major, minor int
	due          time.Time
}

func (d *Data) getActiveMilestones() []string {
	var all []milestone
	for _, dm := range d.Milestones {
		if m := milestoneRE.FindStringSubmatch(*dm.Title); m != nil {
			major, _ := strconv.Atoi(m[1])
			minor, _ := strconv.Atoi(m[3])
			if major <= d.GoReleaseCycle {
				all = append(all, milestone{*dm.Title, major, minor, getTime(dm.DueOn)})
			}
		}
	}
	sort.Sort(milestones(all))
	var titles []string
	for _, m := range all {
		titles = append(titles, m.title)
	}
	return titles
}

type milestones []milestone

func (x milestones) Len() int      { return len(x) }
func (x milestones) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
func (x milestones) Less(i, j int) bool {
	a, b := x[i], x[j]
	if a.major != b.major {
		return a.major < b.major
	}
	if a.minor != b.minor {
		return a.minor < b.minor
	}
	if !a.due.Equal(b.due) {
		return a.due.Before(b.due)
	}
	return a.title < b.title
}

type section struct {
	title string
	count int
	body  string
}

func (d *Data) PrintIssues(w io.Writer) {
	groups := d.groupData(true, false)

	milestones := d.getActiveMilestones()

	sections := []*section{}

	// Issues
	for _, title := range milestones {
		count, body := d.printGroups(groups, false, func(item *Item) bool { return item.Issue != nil && item.Issue.Milestone == title })
		sections = append(sections, &section{
			title, count, body,
		})
	}
	// Pending CLs
	// This uses a different grouping (by CL, not by issue) since
	// otherwise we might print a CL twice.
	count, body := d.printGroups(d.groupData(false, false), false, func(item *Item) bool { return len(item.CLs) > 0 })
	sections = append(sections, &section{
		"Pending CLs",
		count, body,
	})
	// Proposals
	for _, group := range groups {
		if group.Dir == "proposal" {
			count, body := d.printGroups([]*Group{group}, false, func(*Item) bool { return true })
			sections = append(sections, &section{
				"Pending Proposals",
				count, body,
			})
		}
	}
	// Closed
	for _, group := range groups {
		if group.Dir == "closed" {
			count, body := d.printGroups([]*Group{group}, false, func(*Item) bool { return true })
			sections = append(sections, &section{
				"Closed Last Week",
				count, body,
			})
		}
	}

	var titles []string
	for _, s := range sections {
		if s.count > 0 {
			titles = append(titles, fmt.Sprintf("%d %s", s.count, s.title))
		}
	}
	fmt.Fprintf(w, "%s\n", strings.Join(titles, " + "))
	for _, s := range sections {
		if s.count > 0 {
			fmt.Fprintf(w, "\n%s\n%s", s.title, s.body)
		}
	}
}

func (d *Data) PrintCLs(w io.Writer) {
	count, body := d.printGroups(d.groupData(false, true), true, func(item *Item) bool { return len(item.CLs) > 0 })
	fmt.Fprintf(w, "%d Pending CLs\n", count)
	fmt.Fprintf(w, "\n%s\n%s", "Pending CLs", body)
}

func (d *Data) printGroups(groups []*Group, clDetail bool, match func(*Item) bool) (int, string) {
	var output bytes.Buffer
	var count int
	for _, g := range groups {
		if len(groups) != 1 && (g.Dir == "closed" || g.Dir == "proposal") {
			// These groups shouldn't be shown when printing all groups.
			continue
		}
		var header func()
		header = func() {
			if len(groups) > 1 {
				fmt.Fprintf(&output, "\n%s\n", g.Dir)
			}
			header = func() {}
		}
		for _, item := range g.Items {
			if !match(item) {
				continue
			}
			printed := false
			prefix := ""
			if item.Issue != nil {
				header()
				printed = true
				fmt.Fprintf(&output, "    %-10s  %s", fmt.Sprintf("#%d", item.Issue.Number), item.Issue.Title)
				prefix = "\u2937 "
				var tags []string
				if strings.HasSuffix(item.Issue.Milestone, "Early") {
					tags = append(tags, "early")
				}
				if strings.HasSuffix(item.Issue.Milestone, "Maybe") {
					tags = append(tags, "maybe")
				}
				sort.Strings(item.Issue.Labels)
				for _, label := range item.Issue.Labels {
					switch label {
					case "Documentation":
						tags = append(tags, "doc")
					case "Testing":
						tags = append(tags, "test")
					case "Started":
						tags = append(tags, strings.ToLower(label))
					case "Proposal":
						tags = append(tags, "proposal")
					case "Proposal-Accepted":
						tags = append(tags, "proposal-accepted")
					case "Proposal-Declined":
						tags = append(tags, "proposal-declined")
					}
				}
				if len(tags) > 0 {
					fmt.Fprintf(&output, " [%s]", strings.Join(tags, ", "))
				}
				fmt.Fprintf(&output, "\n")
			}
			for _, cl := range item.CLs {
				header()
				printed = true
				fmt.Fprintf(&output, "    %-10s  %s%s\n", fmt.Sprintf("%sCL %d", prefix, cl.Number), prefix, cl.Subject)
				if clDetail {
					fmt.Fprintf(&output, "    %-10s      %s\n", "", cl.Summary(d.Now))
				}
			}
			if printed {
				count++
			}
		}
	}
	return count, output.String()
}
