// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package godash

import (
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"
)

// PrintHTML converts the text output of PrintIssues or PrintCLs into
// an HTML page.
func PrintHTML(w io.Writer, data string) {
	data = html.EscapeString(data)
	i := strings.Index(data, "\n")
	if i < 0 {
		i = len(data)
	}
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<meta charset=\"UTF-8\">\n")
	fmt.Fprintf(w, "<title>%s</title>\n", data[:i])
	fmt.Fprintf(w, "<style>\n")
	fmt.Fprintf(w, ".early {}\n")
	fmt.Fprintf(w, ".maybe {}\n")
	fmt.Fprintf(w, ".late {color: #700; text-decoration: underline;}\n")
	fmt.Fprintf(w, ".closed {background-color: #eee;}\n")
	fmt.Fprintf(w, "hr {border: none; border-top: 2px solid #000; height: 5px; border-bottom: 1px solid #000;}\n")
	fmt.Fprintf(w, "</style>\n")
	fmt.Fprintf(w, "<pre>\n")
	data = regexp.MustCompile(`(?m)^HOWTO`).ReplaceAllString(data, `<a target="_blank" href="/">about the dashboard</a>`)
	data = regexp.MustCompile(`(CL (\d+))\b`).ReplaceAllString(data, "<a target=\"_blank\" href='https://golang.org/cl/$2'>$1</a>")
	data = regexp.MustCompile(`(#(\d\d\d+))\b`).ReplaceAllString(data, "<a target=\"_blank\" href='https://golang.org/issue/$2'>$1</a>")
	data = regexp.MustCompile(`(?m)^(Closed Last Week|Pending Proposals|Pending CLs|Go[\?A-Za-z0-9][^\n]*)`).ReplaceAllString(data, "<hr><b><font size='+1'>$1</font></b>")
	data = regexp.MustCompile(`(?m)^([\?A-Za-z0-9][^\n]*)`).ReplaceAllString(data, "<b>$1</b>")
	data = regexp.MustCompile(`(?m)^([^\n]*\[early[^\n]*)`).ReplaceAllString(data, "<span class='early'>$1</span>")
	data = regexp.MustCompile(`(?m)^([^\n]*\[maybe[^\n]*)`).ReplaceAllString(data, "<span class='maybe'>$1</span>")
	data = regexp.MustCompile(`(?m)^( +)(.*)( → )(.*)(, [\d/]+ days)(, waiting for reviewer)`).ReplaceAllString(data, "$1$2$3<b>$4</b>$5$6")
	data = regexp.MustCompile(`(?m)^( +)(.*)( → )(.*)(, [\d/]+ days)(, waiting for author)`).ReplaceAllString(data, "$1<b>$2</b>$3$4$5$6")
	data = regexp.MustCompile(`(→ )(.*, \d\d+)(/\d+ days)(, waiting for reviewer)`).ReplaceAllString(data, "$1<b class='late'>$2</b>$3$4")
	fmt.Fprintf(w, "%s\n", data)
	fmt.Fprintf(w, "</pre>\n")
}
