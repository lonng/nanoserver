// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"net/http"
	"text/template"

	"golang.org/x/build/dashboard"
)

func handleBuilders(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	if err := buildersTmpl.Execute(&buf, dashboard.Builders); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

var buildersTmpl = template.Must(template.New("builders").Parse(`
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

<h2>Defined Builders</h2>

<table>
<thead><tr><th>name</th><th>pool</th><th>owner</th><th>notes</th></tr>
</thead>
{{range .}}
<tr>
	<td>{{.Name}}</td>
	<td>{{if .IsReverse}}Reverse{{else}}GCE{{end}}</td>
	<td>{{.ShortOwner}}</td>
	<td>{{.Notes}}</td>
</tr>
{{end}}
</table>
</body>
</html>
`))
