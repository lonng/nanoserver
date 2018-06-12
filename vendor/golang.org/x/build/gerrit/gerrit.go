// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gerrit contains code to interact with Gerrit servers.
package gerrit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is a Gerrit client.
type Client struct {
	url  string // URL prefix, e.g. "https://go-review.googlesource.com/a" (without trailing slash)
	auth Auth

	// HTTPClient optionally specifies an HTTP client to use
	// instead of http.DefaultClient.
	HTTPClient *http.Client
}

// NewClient returns a new Gerrit client with the given URL prefix
// and authentication mode.
// The url should be just the scheme and hostname.
// If auth is nil, a default is used, or requests are made unauthenticated.
func NewClient(url string, auth Auth) *Client {
	if auth == nil {
		// TODO(bradfitz): use GitCookies auth, once that exists
		auth = NoAuth
	}
	return &Client{
		url:  strings.TrimSuffix(url, "/"),
		auth: auth,
	}
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

// HTTPError is the error type returned when a Gerrit API call does not return
// the expected status.
type HTTPError struct {
	Res     *http.Response
	Body    []byte // 4KB prefix
	BodyErr error  // any error reading Body
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP status %s; %s", e.Res.Status, e.Body)
}

// doArg is one of urlValues, reqBody, or wantResStatus
type doArg interface {
	isDoArg()
}

type wantResStatus int

func (wantResStatus) isDoArg() {}

type reqBody struct{ body interface{} }

func (reqBody) isDoArg() {}

type urlValues url.Values

func (urlValues) isDoArg() {}

func (c *Client) do(dst interface{}, method, path string, opts ...doArg) error {
	var arg url.Values
	var body interface{}
	var wantStatus = http.StatusOK
	for _, opt := range opts {
		switch opt := opt.(type) {
		case wantResStatus:
			wantStatus = int(opt)
		case reqBody:
			body = opt.body
		case urlValues:
			arg = url.Values(opt)
		default:
			panic(fmt.Sprintf("internal error; unsupported type %T", opt))
		}
	}

	var bodyr io.Reader
	var contentType string
	if body != nil {
		v, err := json.MarshalIndent(body, "", "  ")
		if err != nil {
			return err
		}
		bodyr = bytes.NewReader(v)
		contentType = "application/json"
	}
	// slashA is either "/a" (for authenticated requests) or "" for unauthenticated.
	// See https://gerrit-review.googlesource.com/Documentation/rest-api.html#authentication
	slashA := "/a"
	if _, ok := c.auth.(noAuth); ok {
		slashA = ""
	}
	var err error
	u := c.url + slashA + path
	if arg != nil {
		u += "?" + arg.Encode()
	}
	req, err := http.NewRequest(method, u, bodyr)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	c.auth.setAuth(c, req)
	res, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != wantStatus {
		body, err := ioutil.ReadAll(io.LimitReader(res.Body, 4<<10))
		return &HTTPError{res, body, err}
	}

	// The JSON response begins with an XSRF-defeating header
	// like ")]}\n". Read that and skip it.
	br := bufio.NewReader(res.Body)
	if _, err := br.ReadSlice('\n'); err != nil {
		return err
	}
	return json.NewDecoder(br).Decode(dst)
}

// ChangeInfo is a Gerrit data structure.
// See https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#change-info
type ChangeInfo struct {
	// ID is the ID of the change in the format
	// "'<project>~<branch>~<Change-Id>'", where 'project',
	// 'branch' and 'Change-Id' are URL encoded. For 'branch' the
	// refs/heads/ prefix is omitted.
	ID           string `json:"id"`
	ChangeNumber int    `json:"_number"`

	Project string `json:"project"`

	// Branch is the name of the target branch.
	// The refs/heads/ prefix is omitted.
	Branch string `json:"branch"`

	ChangeID string `json:"change_id"`

	Subject string `json:"subject"`

	// Status is the status of the change (NEW, SUBMITTED, MERGED,
	// ABANDONED, DRAFT).
	Status string `json:"status"`

	Created  TimeStamp `json:"created"`
	Updated  TimeStamp `json:"updated"`
	Mergable bool      `json:"mergable"`

	// CurrentRevision is the commit ID of the current patch set
	// of this change.  This is only set if the current revision
	// is requested or if all revisions are requested (fields
	// "CURRENT_REVISION" or "ALL_REVISIONS").
	CurrentRevision string `json:"current_revision"`

	// Revisions maps a commit ID of the patch set to a
	// RevisionInfo entity.
	//
	// Only set if the current revision is requested (in which
	// case it will only contain a key for the current revision)
	// or if all revisions are requested.
	Revisions map[string]RevisionInfo `json:"revisions"`

	// Owner is the author of the change.
	// The details are only filled in if field "DETAILED_ACCOUNTS" is requested.
	Owner *AccountInfo `json:"owner"`

	// Messages are included if field "MESSAGES" is requested.
	Messages []ChangeMessageInfo `json:"messages"`

	Labels map[string]LabelInfo `json:"labels"`

	// TODO: more as needed

	// MoreChanges is set on the last change from QueryChanges if
	// the result set is truncated by an 'n' parameter.
	MoreChanges bool `json:"_more_changes"`
}

type AccountInfo struct {
	NumericID int64  `json:"_account_id"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Username  string `json:"username,omitempty"`
}

func (ai *AccountInfo) Equal(v *AccountInfo) bool {
	if ai == nil || v == nil {
		return false
	}
	return ai.NumericID == v.NumericID
}

type ChangeMessageInfo struct {
	ID             string       `json:"id"`
	Author         *AccountInfo `json:"author"`
	Time           TimeStamp    `json:"date"`
	Message        string       `json:"message"`
	RevisionNumber int          `json:"_revision_number"`
}

// The LabelInfo entity contains information about a label on a
// change, always corresponding to the current patch set.
//
// There are two options that control the contents of LabelInfo:
// LABELS and DETAILED_LABELS.
//
// For a quick summary of the state of labels, use LABELS.
//
// For detailed information about labels, including exact numeric
// votes for all users and the allowed range of votes for the current
// user, use DETAILED_LABELS.
type LabelInfo struct {
	// Optional means the label may be set, but it’s neither
	// necessary for submission nor does it block submission if
	// set.
	Optional bool `json:"optional"`

	// Fields set by LABELS field option:

	All []ApprovalInfo `json:"all"`
}

type ApprovalInfo struct {
	AccountInfo
	Value int       `json:"value"`
	Date  TimeStamp `json:"date"`
}

// The RevisionInfo entity contains information about a patch set. Not
// all fields are returned by default. Additional fields can be
// obtained by adding o parameters as described at:
// https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#list-changes
type RevisionInfo struct {
	Draft          bool                  `json:"draft"`
	PatchSetNumber int                   `json:"_number"`
	Created        TimeStamp             `json:"created"`
	Uploader       *AccountInfo          `json:"uploader"`
	Ref            string                `json:"ref"`
	Fetch          map[string]*FetchInfo `json:"fetch"`
	Commit         *CommitInfo           `json:"commit"`
	Files          map[string]*FileInfo  `json:"files"`
	// TODO: more
}

type CommitInfo struct {
	Author    GitPersonInfo `json:"author"`
	Committer GitPersonInfo `json:"committer"`
	CommitID  string        `json:"commit"`
	Subject   string        `json:"subject"`
	Message   string        `json:"message"`
	Parents   []CommitInfo  `json:"parents"`
}

type GitPersonInfo struct {
	Name     string    `json:"name"`
	Email    string    `json:"Email"`
	Date     TimeStamp `json:"date"`
	TZOffset int       `json:"tz"`
}

type FileInfo struct {
	Status        string `json:"status"`
	Binary        bool   `json:"binary"`
	OldPath       string `json:"old_path"`
	LinesInserted int    `json:"lines_inserted"`
	LinesDeleted  int    `json:"lines_deleted"`
}

type FetchInfo struct {
	URL      string            `json:"url"`
	Ref      string            `json:"ref"`
	Commands map[string]string `json:"commands"`
}

// QueryChangesOpt are options for QueryChanges.
type QueryChangesOpt struct {
	// N is the number of results to return.
	// If 0, the 'n' parameter is not sent to Gerrit.
	N int

	// Fields are optional fields to also return.
	// Example strings include "ALL_REVISIONS", "LABELS", "MESSAGES".
	// For a complete list, see:
	// https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#change-info
	Fields []string
}

func condInt(n int) []string {
	if n != 0 {
		return []string{strconv.Itoa(n)}
	}
	return nil
}

// QueryChanges queries changes. The q parameter is a Gerrit search query.
// For the API call, see https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#list-changes
// For the query syntax, see https://gerrit-review.googlesource.com/Documentation/user-search.html#_search_operators
func (c *Client) QueryChanges(q string, opts ...QueryChangesOpt) ([]*ChangeInfo, error) {
	var opt QueryChangesOpt
	switch len(opts) {
	case 0:
	case 1:
		opt = opts[0]
	default:
		return nil, errors.New("only 1 option struct supported")
	}
	var changes []*ChangeInfo
	err := c.do(&changes, "GET", "/changes/", urlValues{
		"q": {q},
		"n": condInt(opt.N),
		"o": opt.Fields,
	})
	return changes, err
}

// GetChangeDetail retrieves a change with labels, detailed labels, detailed
// accounts, and messages.
// For the API call, see https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#get-change-detail
func (c *Client) GetChangeDetail(changeID string) (*ChangeInfo, error) {
	var change ChangeInfo
	err := c.do(&change, "GET", "/changes/"+changeID+"/detail")
	if err != nil {
		return nil, err
	}
	return &change, nil
}

type ReviewInput struct {
	Message string         `json:"message,omitempty"`
	Labels  map[string]int `json:"labels,omitempty"`
}

type reviewInfo struct {
	Labels map[string]int `json:"labels,omitempty"`
}

// SetReview leaves a message on a change and/or modifies labels.
// For the API call, see https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#set-review
// The changeID is https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#change-id
// The revision is https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#revision-id
func (c *Client) SetReview(changeID, revision string, review ReviewInput) error {
	var res reviewInfo
	return c.do(&res, "POST", fmt.Sprintf("/changes/%s/revisions/%s/review", changeID, revision),
		reqBody{review})
}

// AbandonChange abandons the given change.
func (c *Client) AbandonChange(changeID string) error {
	var change ChangeInfo
	return c.do(&change, "POST", "/changes/"+changeID+"/abandon")
}

// ProjectInput contains the options for creating a new project.
// See https://gerrit-review.googlesource.com/Documentation/rest-api-projects.html#project-input
type ProjectInput struct {
	Parent      string `json:"parent,omitempty"`
	Description string `json:"description,omitempty"`
	SubmitType  string `json:"submit_type,omitempty"`

	CreateNewChangeForAllNotInTarget string `json:"create_new_change_for_all_not_in_target,omitempty"`

	// TODO(bradfitz): more, as needed.
}

// ProjectInfo is information about a Gerrit project.
// See https://gerrit-review.googlesource.com/Documentation/rest-api-projects.html#project-info
type ProjectInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Parent      string            `json:"parent"`
	Description string            `json:"description"`
	State       string            `json:"state"`
	Branches    map[string]string `json:"branches"`
}

// CreateProject creates a new project.
func (c *Client) CreateProject(name string, p ...ProjectInput) (ProjectInfo, error) {
	var pi ProjectInput
	if len(p) > 1 {
		panic("invalid use of multiple project inputs")
	}
	if len(p) == 1 {
		pi = p[0]
	}
	var res ProjectInfo
	err := c.do(&res, "PUT", fmt.Sprintf("/projects/%s", name), reqBody{&pi}, wantResStatus(http.StatusCreated))
	return res, err
}

// ErrProjectNotExist is returned when a project doesn't exist.
// It is not necessarily returned unless a method is documented as
// returning it.
var ErrProjectNotExist = errors.New("gerrit: requested project does not exist")

// GetProjectInfo returns info about a project.
// If the project doesn't exist, the error will be ErrProjectNotExist.
func (c *Client) GetProjectInfo(name string) (ProjectInfo, error) {
	var res ProjectInfo
	err := c.do(&res, "GET", fmt.Sprintf("/projects/%s", name))
	if he, ok := err.(*HTTPError); ok && he.Res.StatusCode == 404 {
		return res, ErrProjectNotExist
	}
	return res, err
}

// BranchInfo is information about a branch.
// See https://gerrit-review.googlesource.com/Documentation/rest-api-projects.html#branch-info
type BranchInfo struct {
	Ref       string `json:"ref"`
	Revision  string `json:"revision"`
	CanDelete bool   `json:"can_delete"`
}

// GetProjectBranches returns a project's branches.
func (c *Client) GetProjectBranches(name string) (map[string]BranchInfo, error) {
	var res []BranchInfo
	err := c.do(&res, "GET", fmt.Sprintf("/projects/%s/branches/", name))
	if err != nil {
		return nil, err
	}
	m := map[string]BranchInfo{}
	for _, bi := range res {
		m[bi.Ref] = bi
	}
	return m, nil
}

// GetAccountInfo gets the specified account's information from Gerrit.
// For the API call, see https://gerrit-review.googlesource.com/Documentation/rest-api-accounts.html#get-account
// The accountID is https://gerrit-review.googlesource.com/Documentation/rest-api-accounts.html#account-id
//
// Note that getting "self" is a good way to validate host access, since it only requires peeker
// access to the host, not to any particular repository.
func (c *Client) GetAccountInfo(accountID string) (AccountInfo, error) {
	var res AccountInfo
	err := c.do(&res, "GET", fmt.Sprintf("/accounts/%s", accountID))
	return res, err
}

type TimeStamp time.Time

// Gerrit's timestamp layout is like time.RFC3339Nano, but with a space instead of the "T",
// and without a timezone (it's always in UTC).
const timeStampLayout = "2006-01-02 15:04:05.999999999"

func (ts *TimeStamp) UnmarshalJSON(p []byte) error {
	if len(p) < 2 {
		return errors.New("Timestamp too short")
	}
	if p[0] != '"' || p[len(p)-1] != '"' {
		return errors.New("not double-quoted")
	}
	s := strings.Trim(string(p), "\"")
	t, err := time.Parse(timeStampLayout, s)
	if err != nil {
		return err
	}
	*ts = TimeStamp(t)
	return nil
}

func (ts TimeStamp) Time() time.Time { return time.Time(ts) }
