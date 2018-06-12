// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Code related to remote buildlets. See x/build/remote-buildlet.txt

package main // import "golang.org/x/build/cmd/coordinator"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/http/httputil"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/build/buildlet"
	"golang.org/x/build/dashboard"
	"golang.org/x/net/context"
)

var (
	remoteBuildlets = struct {
		sync.Mutex
		m map[string]*remoteBuildlet // keyed by buildletName
	}{m: map[string]*remoteBuildlet{}}

	cleanTimer *time.Timer
)

const (
	remoteBuildletIdleTimeout   = 30 * time.Minute
	remoteBuildletCleanInterval = time.Minute
)

func init() {
	cleanTimer = time.AfterFunc(remoteBuildletCleanInterval, expireBuildlets)
}

type remoteBuildlet struct {
	User    string // "user-foo" build key
	Name    string // dup of key
	Type    string
	Created time.Time
	Expires time.Time

	buildlet *buildlet.Client
}

func addRemoteBuildlet(rb *remoteBuildlet) (name string) {
	remoteBuildlets.Lock()
	defer remoteBuildlets.Unlock()
	n := 0
	for {
		name = fmt.Sprintf("%s-%s-%d", rb.User, rb.Type, n)
		if _, ok := remoteBuildlets.m[name]; ok {
			n++
		} else {
			remoteBuildlets.m[name] = rb
			return name
		}
	}
}

func expireBuildlets() {
	defer cleanTimer.Reset(remoteBuildletCleanInterval)
	remoteBuildlets.Lock()
	defer remoteBuildlets.Unlock()
	now := time.Now()
	for name, rb := range remoteBuildlets.m {
		if !rb.Expires.IsZero() && rb.Expires.Before(now) {
			go rb.buildlet.Close()
			delete(remoteBuildlets.m, name)
		}
	}
}

// always wrapped in requireBuildletProxyAuth.
func handleBuildletCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST required", 400)
		return
	}
	typ := r.FormValue("type")
	if typ == "" {
		http.Error(w, "missing 'type' parameter", 400)
		return
	}
	bconf, ok := dashboard.Builders[typ]
	if !ok {
		http.Error(w, "unknown builder type in 'type' parameter", 400)
		return
	}
	user, _, _ := r.BasicAuth()
	pool := poolForConf(bconf)

	var closeNotify <-chan bool
	if cn, ok := w.(http.CloseNotifier); ok {
		closeNotify = cn.CloseNotify()
	}

	ctx := context.WithValue(context.Background(), buildletTimeoutOpt{}, time.Duration(0))
	ctx, cancel := context.WithCancel(ctx)

	// Doing a release?
	if user == "release" || user == "adg" || user == "bradfitz" {
		ctx = context.WithValue(ctx, highPriorityOpt{}, true)
	}

	resc := make(chan *buildlet.Client)
	errc := make(chan error)
	go func() {
		bc, err := pool.GetBuildlet(ctx, typ, loggerFunc(func(event string, optText ...string) {
			var extra string
			if len(optText) > 0 {
				extra = " " + optText[0]
			}
			log.Printf("creating buildlet %s for %s: %s%s", typ, user, event, extra)
		}))
		if bc != nil {
			resc <- bc
			return
		}
		errc <- err
	}()
	for {
		select {
		case bc := <-resc:
			rb := &remoteBuildlet{
				User:     user,
				Type:     typ,
				buildlet: bc,
				Created:  time.Now(),
				Expires:  time.Now().Add(remoteBuildletIdleTimeout),
			}
			rb.Name = addRemoteBuildlet(rb)
			jenc, err := json.MarshalIndent(rb, "", "  ")
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Print(err)
				return
			}
			log.Printf("created buildlet %v for %v", rb.Name, rb.User)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			jenc = append(jenc, '\n')
			w.Write(jenc)
			return
		case err := <-errc:
			log.Printf("error creating buildlet: %v", err)
			http.Error(w, err.Error(), 500)
			return
		case <-closeNotify:
			log.Printf("client went away during buildlet create request")
			cancel()
			closeNotify = nil // unnecessary, but habit.
		}
	}
}

// always wrapped in requireBuildletProxyAuth.
func handleBuildletList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "POST required", 400)
		return
	}
	res := make([]*remoteBuildlet, 0) // so it's never JSON "null"
	remoteBuildlets.Lock()
	defer remoteBuildlets.Unlock()
	user, _, _ := r.BasicAuth()
	for _, rb := range remoteBuildlets.m {
		if rb.User == user {
			res = append(res, rb)
		}
	}
	sort.Sort(byBuildletName(res))
	jenc, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jenc = append(jenc, '\n')
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jenc)
}

type byBuildletName []*remoteBuildlet

func (s byBuildletName) Len() int           { return len(s) }
func (s byBuildletName) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s byBuildletName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func remoteBuildletStatus() string {
	remoteBuildlets.Lock()
	defer remoteBuildlets.Unlock()

	if len(remoteBuildlets.m) == 0 {
		return "<i>(none)</i>"
	}

	var buf bytes.Buffer
	var all []*remoteBuildlet
	for _, rb := range remoteBuildlets.m {
		all = append(all, rb)
	}
	sort.Sort(byBuildletName(all))

	buf.WriteString("<ul>")
	for _, rb := range all {
		fmt.Fprintf(&buf, "<li><b>%s</b>, created %v ago, expires in %v</li>\n",
			html.EscapeString(rb.Name),
			time.Since(rb.Created), rb.Expires.Sub(time.Now()))
	}
	buf.WriteString("</ul>")

	return buf.String()
}

// httpRouter separates out HTTP traffic being proxied
// to buildlets on behalf of remote clients from traffic
// destined for the coordiantor itself (the default).
type httpRouter struct{}

func (httpRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Buildlet-Proxy") != "" {
		requireBuildletProxyAuth(http.HandlerFunc(proxyBuildletHTTP)).ServeHTTP(w, r)
	} else {
		http.DefaultServeMux.ServeHTTP(w, r)
	}
}

func proxyBuildletHTTP(w http.ResponseWriter, r *http.Request) {
	if r.TLS == nil {
		http.Error(w, "https required", http.StatusBadRequest)
		return
	}
	buildletName := r.Header.Get("X-Buildlet-Proxy")
	if buildletName == "" {
		http.Error(w, "missing X-Buildlet-Proxy; server misconfig", http.StatusInternalServerError)
		return
	}
	remoteBuildlets.Lock()
	rb, ok := remoteBuildlets.m[buildletName]
	if ok {
		rb.Expires = time.Now().Add(remoteBuildletIdleTimeout)
	}
	remoteBuildlets.Unlock()
	if !ok {
		http.Error(w, "unknown or expired buildlet", http.StatusBadGateway)
		return
	}
	user, _, _ := r.BasicAuth()
	if rb.User != user {
		http.Error(w, "you don't own that buildlet", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" && r.URL.Path == "/halt" {
		err := rb.buildlet.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		rb.buildlet.Close()
		remoteBuildlets.Lock()
		delete(remoteBuildlets.m, buildletName)
		remoteBuildlets.Unlock()
		return
	}

	outReq, err := http.NewRequest(r.Method, rb.buildlet.URL()+r.URL.Path+"?"+r.URL.RawQuery, r.Body)
	if err != nil {
		log.Printf("bad proxy request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	outReq.Header = r.Header
	proxy := &httputil.ReverseProxy{
		Director:      func(*http.Request) {}, // nothing
		Transport:     rb.buildlet.ProxyRoundTripper(),
		FlushInterval: 500 * time.Millisecond,
	}
	proxy.ServeHTTP(w, outReq)
}

func requireBuildletProxyAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "missing required authentication", 400)
			return
		}
		if !strings.HasPrefix(user, "user-") || builderKey(user) != pass {
			http.Error(w, "bad username or password", 401)
			return
		}
		h.ServeHTTP(w, r)
	})
}
