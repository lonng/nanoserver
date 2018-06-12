// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package godash

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/build/gerrit"
)

// Parsing of Gerrit information and review messages to produce CL structure.

var (
	reviewerRE     = regexp.MustCompile(`(?m)^R=([\w\-.@]+)\b`)
	scoreRE        = regexp.MustCompile(`\APatch Set \d+: Code-Review([+-][0-9]+)`)
	removedScoreRE = regexp.MustCompile(`\ARemoved the following votes:\n\n\* Code-Review([+-][0-9]+) by [^\n]* <([^\n]*)>`)
	issueRefRE     = regexp.MustCompile(`#\d+\b`)
	goIssueRefRE   = regexp.MustCompile(`\bgolang/go#\d+\b`)
)

// ParseCL takes a ChangeInfo as returned from the Gerrit API and
// applies Go-project-specific logic to turn it into a CL struct. The
// primary information that is added is the CL's state in the Go
// review process, based on parsing R= lines in comments on the CL.
func ParseCL(ci *gerrit.ChangeInfo, reviewers *Reviewers, goReleaseCycle int) *CL {
	// Gather information.
	var (
		scores           = make(map[string]int)
		initialReviewer  = ""
		firstResponder   = ""
		explicitReviewer = ""
		closeReason      = ""
	)
	for _, msg := range ci.Messages {
		if msg.Author == nil { // happens for Gerrit-generated messages
			continue
		}
		if strings.HasPrefix(msg.Message, "Uploaded patch set ") {
			if explicitReviewer == "close" && !strings.HasPrefix(closeReason, "Go") {
				explicitReviewer = ""
				closeReason = ""
			}
			for who, score := range scores {
				if score == +1 || score == -1 {
					delete(scores, who)
				}
			}
		}
		if m := reviewerRE.FindStringSubmatch(msg.Message); m != nil {
			if m[1] == "close" {
				explicitReviewer = "close"
				closeReason = "Closed"
			} else if strings.HasPrefix(m[1], "go1.") {
				n, _ := strconv.Atoi(m[1][len("go1."):])
				if n > goReleaseCycle {
					explicitReviewer = "close"
					closeReason = "Go" + m[1][2:]
				}
			} else if m[1] == "golang-dev" || m[1] == "golang-codereviews" {
				explicitReviewer = "golang-dev"
			} else if x := reviewers.Resolve(m[1]); x != "" {
				explicitReviewer = x
			}
		}
		if m := scoreRE.FindStringSubmatch(msg.Message); m != nil {
			n, _ := strconv.Atoi(m[1])
			scores[msg.Author.Email] = n
		}
		if m := removedScoreRE.FindStringSubmatch(msg.Message); m != nil {
			delete(scores, m[1])
		}
		if firstResponder == "" && reviewers.IsReviewer(msg.Author.Email) && msg.Author.Email != ci.Owner.Email {
			firstResponder = msg.Author.Email
		}
	}

	cl := &CL{
		Number:      ci.ChangeNumber,
		Subject:     ci.Subject,
		Project:     ci.Project,
		Author:      reviewers.Shorten(ci.Owner.Email),
		AuthorEmail: ci.Owner.Email,
		Scores:      scores,
		Status:      strings.ToLower(ci.Status),
	}

	// Determine reviewer, in priorty order.
	// When breaking ties, give preference to R= setting.
	// Otherwise compare by email address.
	maybe := func(who string) {
		if cl.ReviewerEmail == "" || who == explicitReviewer || cl.ReviewerEmail != explicitReviewer && cl.ReviewerEmail > who {
			cl.ReviewerEmail = who
		}
	}

	// 1. Anyone who -2'ed the CL.
	if cl.ReviewerEmail == "" {
		for who, score := range scores {
			if score == -2 {
				maybe(who)
			}
		}
	}

	// 2. Anyone who +2'ed the CL.
	if cl.ReviewerEmail == "" {
		for who, score := range scores {
			if score == +2 {
				maybe(who)
			}
		}
	}

	// 2½. Even if a CL is +2 or -2, R=closed wins,
	// so that it doesn't appear in listings by default.
	if explicitReviewer == "close" {
		cl.ReviewerEmail = "close"
	}

	// 3. Last explicit R= in review message.
	if cl.ReviewerEmail == "" {
		cl.ReviewerEmail = explicitReviewer
	}
	// 4. Initial target of review requecl.
	// TODO: If there's some way to figure this out, do so.
	_ = initialReviewer
	// 5. Whoever responds first and looks like a reviewer.
	if cl.ReviewerEmail == "" {
		cl.ReviewerEmail = firstResponder
	}

	// Allow R=golang-dev in #2 as "unassign".
	if cl.ReviewerEmail == "golang-dev" {
		cl.ReviewerEmail = ""
	}

	cl.Reviewer = reviewers.Shorten(cl.ReviewerEmail)

	// Now that we know who the reviewer is,
	// figure out whether the CL is in need of review
	// (or else is waiting for the author to do more work).
	for _, msg := range ci.Messages {
		if msg.Author == nil { // happens for Gerrit-generated messages
			continue
		}
		if cl.Start.IsZero() {
			cl.Start = msg.Time.Time()
		}
		if strings.HasPrefix(msg.Message, "Uploaded patch set ") {
			cl.NeedsReview = true
			cl.NeedsReviewChanged = msg.Time.Time()
		}
		if msg.Author.Email == cl.ReviewerEmail {
			cl.NeedsReview = false
			cl.NeedsReviewChanged = msg.Time.Time()
		}
	}

	if cl.ReviewerEmail == "close" {
		cl.Reviewer = closeReason
		cl.ReviewerEmail = ""
		cl.Closed = true
		cl.NeedsReview = false
	}

	// DO NOT REVIEW overrides anything in the CL state.
	// We shouldn't see these, because the query always
	// contains -message:do-not-review, but check anyway.
	if _, ok := ci.Labels["Do-Not-Review"]; ok {
		cl.Reviewer = "DoNotReview"
		cl.ReviewerEmail = ""
		cl.DoNotReview = true
		cl.NeedsReview = false
	}

	// Find issue numbers.
	cl.Issues = []int{} // non-nil for json
	refRE := issueRefRE
	if cl.Project != "go" {
		refRE = goIssueRefRE
	}

	rev := ci.Revisions[ci.CurrentRevision]
	for file := range rev.Files {
		cl.Files = append(cl.Files, file)
	}
	sort.Strings(cl.Files)
	if rev.Commit != nil {
		for _, ref := range refRE.FindAllString(rev.Commit.Message, -1) {
			n, _ := strconv.Atoi(ref[strings.Index(ref, "#")+1:])
			if n != 0 {
				cl.Issues = append(cl.Issues, n)
			}
		}
	}
	cl.Issues = uniq(cl.Issues)

	return cl
}

func uniq(x []int) []int {
	sort.Ints(x)
	out := x[:0]
	for _, v := range x {
		if len(out) == 0 || out[len(out)-1] != v {
			out = append(out, v)
		}
	}
	return out
}

// CL records information about a single CL.
// This is also used by golang.org/x/build/cmd/cl and any changes need
// to reflected in its doc comment.
type CL struct {
	Number             int            // CL number
	Subject            string         // subject (first line of commit message)
	Project            string         // "go" or a subrepository name
	Author             string         // author, short form or else full email
	AuthorEmail        string         // author, full email
	Reviewer           string         // expected reviewer, short form or else full email
	ReviewerEmail      string         // expected reviewer, full email
	Start              time.Time      // time CL was first uploaded
	NeedsReview        bool           // CL is waiting for reviewer (otherwise author)
	NeedsReviewChanged time.Time      // time NeedsReview last changed
	Closed             bool           // CL closed with R=close
	DoNotReview        bool           // CL marked DO NOT REVIEW
	Issues             []int          // issues referenced by commit message
	Scores             map[string]int // current review scores
	Files              []string       // files changed in CL
	Status             string         // "new", "submitted", "merged", ...
}

func (cl *CL) Age(now time.Time) time.Duration {
	return now.Sub(cl.Start)
}

func (cl *CL) Delay(now time.Time) time.Duration {
	return now.Sub(cl.NeedsReviewChanged)
}

func (cl *CL) Summary(now time.Time) string {
	var buf bytes.Buffer
	who := "author"
	if cl.NeedsReview {
		who = "reviewer"
	}
	rev := cl.Reviewer
	if rev == "" {
		rev = "???"
	}
	score := ""
	if x := cl.Scores[cl.ReviewerEmail]; x != 0 {
		score = fmt.Sprintf("%+d", x)
	}
	fmt.Fprintf(&buf, "%s → %s%s, %d/%d days, waiting for %s", cl.Author, rev, score, int(now.Sub(cl.NeedsReviewChanged).Seconds()/86400), int(now.Sub(cl.Start).Seconds()/86400), who)
	for _, id := range cl.Issues {
		fmt.Fprintf(&buf, " #%d", id)
	}
	return buf.String()
}
