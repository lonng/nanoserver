// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package devapp implements a simple App Engine app for generating
// and serving Go project release dashboards using the godash
// command/library.
package devapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/build/gerrit"
	"golang.org/x/build/godash"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

const entityPrefix = "DevApp"

func init() {
	for _, page := range []string{"release", "cl"} {
		page := page
		http.Handle("/"+page, hstsHandler(func(w http.ResponseWriter, r *http.Request) { servePage(w, r, page) }))
	}
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/setToken", setTokenHandler)
}

// hstsHandler wraps an http.HandlerFunc such that it sets the HSTS header.
func hstsHandler(fn http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; preload")
		fn(w, r)
	})
}

type Page struct {
	// Content is the complete HTML of the page.
	Content []byte
}

func servePage(w http.ResponseWriter, r *http.Request, page string) {
	ctx := appengine.NewContext(r)
	var entity Page
	if err := datastore.Get(ctx, datastore.NewKey(ctx, entityPrefix+"Page", page, 0, nil), &entity); err != nil {
		http.Error(w, "page not found", 404)
		return
	}
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	w.Write(entity.Content)
}

type Cache struct {
	Value []byte
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if err := update(ctx); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func writePage(ctx context.Context, page string, content []byte) error {
	entity := &Page{
		Content: content,
	}
	_, err := datastore.Put(ctx, datastore.NewKey(ctx, entityPrefix+"Page", page, 0, nil), entity)
	return err
}

func update(ctx context.Context) error {
	var token, cache Cache
	var keys []*datastore.Key
	keys = append(keys, datastore.NewKey(ctx, entityPrefix+"Cache", "github-token", 0, nil))
	keys = append(keys, datastore.NewKey(ctx, entityPrefix+"Cache", "reviewers", 0, nil))
	datastore.GetMulti(ctx, keys, []*Cache{&token, &cache}) // Ignore errors since they might not exist.
	// Without a deadline, urlfetch will use a 5s timeout which is too slow for Gerrit.
	ctx, cancel := context.WithTimeout(ctx, 9*time.Minute)
	defer cancel()
	transport := &urlfetch.Transport{Context: ctx}
	gh := godash.NewGitHubClient("golang/go", string(token.Value), transport)
	ger := gerrit.NewClient("https://go-review.googlesource.com", gerrit.NoAuth)
	ger.HTTPClient = urlfetch.Client(ctx)
	data := &godash.Data{Reviewers: &godash.Reviewers{}}
	if len(cache.Value) > 0 {
		if err := json.Unmarshal(cache.Value, data.Reviewers); err != nil {
			return err
		}
	}

	if err := data.Reviewers.LoadGithub(gh); err != nil {
		return err
	}
	if err := data.FetchData(gh, ger, 7, false, false); err != nil {
		log.Criticalf(ctx, "failed to fetch data: %v", err)
		return err
	}

	for _, cls := range []bool{false, true} {
		var output bytes.Buffer
		kind := "release"
		if cls {
			kind = "CL"
		}
		fmt.Fprintf(&output, "Go %s dashboard\n", kind)
		fmt.Fprintf(&output, "%v\n\n", time.Now().UTC().Format(time.UnixDate))
		fmt.Fprintf(&output, "HOWTO\n\n")
		if cls {
			data.PrintCLs(&output)
		} else {
			data.PrintIssues(&output)
		}
		var html bytes.Buffer
		godash.PrintHTML(&html, output.String())

		if err := writePage(ctx, strings.ToLower(kind), html.Bytes()); err != nil {
			return err
		}
	}

	// N.B. We can't serialize the whole data because a) it's too
	// big and b) it can only be serialized by Go >=1.7.
	js, err := json.MarshalIndent(data.Reviewers, "", "  ")
	if err != nil {
		return err
	}
	cache.Value = js
	if _, err = datastore.Put(ctx, datastore.NewKey(ctx, entityPrefix+"Cache", "reviewers", 0, nil), &cache); err != nil {
		return err
	}
	return nil
}

func setTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	r.ParseForm()
	if value := r.Form.Get("value"); value != "" {
		var token Cache
		token.Value = []byte(value)
		if _, err := datastore.Put(ctx, datastore.NewKey(ctx, entityPrefix+"Cache", "github-token", 0, nil), &token); err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}
