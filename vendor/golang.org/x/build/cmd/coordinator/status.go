// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"sort"
	"strings"
	"time"
)

func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	round := func(t time.Duration) time.Duration {
		return t / time.Second * time.Second
	}
	df := diskFree()

	statusMu.Lock()
	data := statusData{
		Total:    len(status),
		Uptime:   round(time.Now().Sub(processStartTime)),
		Recent:   append([]*buildStatus{}, statusDone...),
		DiskFree: df,
		Version:  Version,
	}
	for _, st := range status {
		data.Active = append(data.Active, st)
	}
	// TODO: make this prettier.
	var buf bytes.Buffer
	for _, key := range tryList {
		if ts := tries[key]; ts != nil {
			state := ts.state()
			fmt.Fprintf(&buf, "Change-ID: %v Commit: %v (<a href='/try?commit=%v'>status</a>)\n",
				key.ChangeTriple(), key.Commit, key.Commit[:8])
			fmt.Fprintf(&buf, "   Remain: %d, fails: %v\n", state.remain, state.failed)
			for _, bs := range ts.builds {
				fmt.Fprintf(&buf, "  %s: running=%v\n", bs.name, bs.isRunning())
			}
		}
	}
	statusMu.Unlock()

	data.RemoteBuildlets = template.HTML(remoteBuildletStatus())

	sort.Sort(byAge(data.Active))
	sort.Sort(sort.Reverse(byAge(data.Recent)))
	if errTryDeps != nil {
		data.TrybotsErr = errTryDeps.Error()
	} else {
		if buf.Len() == 0 {
			data.Trybots = template.HTML("<i>(none)</i>")
		} else {
			data.Trybots = template.HTML("<pre>" + buf.String() + "</pre>")
		}
	}

	buf.Reset()
	gcePool.WriteHTMLStatus(&buf)
	data.GCEPoolStatus = template.HTML(buf.String())
	buf.Reset()

	kubePool.WriteHTMLStatus(&buf)
	data.KubePoolStatus = template.HTML(buf.String())
	buf.Reset()

	reversePool.WriteHTMLStatus(&buf)
	data.ReversePoolStatus = template.HTML(buf.String())

	buf.Reset()
	if err := statusTmpl.Execute(&buf, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

func diskFree() string {
	out, _ := exec.Command("df", "-h").Output()
	return string(out)
}

// statusData is the data that fills out statusTmpl.
type statusData struct {
	Total             int
	Uptime            time.Duration
	Active            []*buildStatus
	Recent            []*buildStatus
	TrybotsErr        string
	Trybots           template.HTML
	GCEPoolStatus     template.HTML // TODO: embed template
	KubePoolStatus    template.HTML // TODO: embed template
	ReversePoolStatus template.HTML // TODO: embed template
	RemoteBuildlets   template.HTML
	DiskFree          string
	Version           string
}

var statusTmpl = template.Must(template.New("status").Parse(`
<!DOCTYPE html>
<html>
<head><link rel="stylesheet" href="/style.css"/><title>Go Farmer</title></head>
<body>
<header>
	<h1>Go Build Coordinator</h1>
	<nav>
		<a href="https://build.golang.org">Dashboard</a>
		<a href="/builders">Builders</a>
	</nav>
	<div class="clear"></div>
</header>

<h2>Running</h2>
<p>{{printf "%d" .Total}} total builds active. Uptime {{printf "%s" .Uptime}}. Version {{.Version}}.

<h2 id=trybots><a href='#trybots'>🔗</a> Active Trybot Runs</h2>
{{- if .TrybotsErr}}
<b>trybots disabled:</b>: {{.TrybotsErr}}
{{else}}
{{.Trybots}}
{{end}}

<h2 id=remote><a href='#remote'>🔗</a> Remote buildlets</h3>
{{.RemoteBuildlets}}

<h2 id=pools><a href='#pools'>🔗</a> Buildlet pools</h2>
<ul>
<li>{{.GCEPoolStatus}}</li>
<li>{{.KubePoolStatus}}</li>
<li>{{.ReversePoolStatus}}</li>
</ul>

<h2 id=active><a href='#active'>🔗</a> Active builds</h2>
<ul>
{{range .Active}}
<li><pre>{{.HTMLStatusLine}}</pre></li>
{{end}}
</ul>

<h2 id=completed><a href='#completed'>🔗</a> Recently completed</h2>
<ul>
{{range .Recent}}
<li><span>{{.HTMLStatusLine_done}}</span></li>
{{end}}
</ul>

<h2 id=disk><a href='#disk'>🔗</a> Disk Space</h2>
<pre>{{.DiskFree}}</pre>

</body>
</html>
`))

func handleStyleCSS(w http.ResponseWriter, r *http.Request) {
	src := strings.NewReader(styleCSS)
	http.ServeContent(w, r, "style.css", processStartTime, src)
}

const styleCSS = `
body {
	font-family: sans-serif;
	color: #222;
	padding: 10px;
	margin: 0;
}

h1, h2 { color: #375EAB; }
h1 { font-size: 24px; }
h2 { font-size: 20px; }

pre {
	font-family: monospace;
	font-size: 9pt;
}

header {
	margin: -10px -10px 0 -10px;
	padding: 10px 10px;
	background: #E0EBF5;
}
header a { color: #222; }
header h1 {
	display: inline;
	margin: 0;
	padding-top: 5px;
}
header nav {
	display: inline-block;
	margin-left: 20px;
}
header nav a {
	display: inline-block;
	padding: 10px;
	margin: 0;
	margin-right: 5px;
	color: white;
	background: #375EAB;
	text-decoration: none;
	font-size: 16px;
	border: 1px solid #375EAB;
	border-radius: 5px;
}

table {
	border-collapse: collapse;
	font-size: 9pt;
}

table td, table th, table td, table th {
	text-align: left;
	vertical-align: top;
	padding: 2px 6px;
}

table thead tr {
	background: #fff !important;
}
`
