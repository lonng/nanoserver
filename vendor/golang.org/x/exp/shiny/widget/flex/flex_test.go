// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flex

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"testing"

	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
)

type layoutTest struct {
	direction    Direction
	wrap         FlexWrap
	alignContent AlignContent
	size         image.Point       // size of container
	measured     [][2]float64      // MeasuredSize of child elements
	layoutData   []LayoutData      // LayoutData of child elements
	want         []image.Rectangle // final Rect of child elements
}

func (t *layoutTest) html() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, `<style>
#container {
	display: flex;
	width:   %dpx;
	height:  %dpx;
`, t.size.X, t.size.Y)

	switch t.direction {
	case Row:
	case RowReverse:
		fmt.Fprintf(buf, "\tflex-direction: row-reverse;\n")
	case Column:
		fmt.Fprintf(buf, "\tflex-direction: column;\n")
	case ColumnReverse:
		fmt.Fprintf(buf, "\tflex-direction: column-reverse;\n")
	}
	switch t.wrap {
	case NoWrap:
	case Wrap:
		fmt.Fprintf(buf, "\tflex-wrap: wrap;\n")
	case WrapReverse:
		fmt.Fprintf(buf, "\tflex-wrap: wrap-reverse;\n")
	}
	switch t.alignContent {
	case AlignContentStart:
	case AlignContentEnd:
		fmt.Fprintf(buf, "\talign-content: flex-end;\n")
	case AlignContentCenter:
		fmt.Fprintf(buf, "\talign-content: center;\n")
	case AlignContentSpaceBetween:
		fmt.Fprintf(buf, "\talign-content: space-between;\n")
	case AlignContentSpaceAround:
		fmt.Fprintf(buf, "\talign-content: space-around;\n")
	case AlignContentStretch:
		fmt.Fprintf(buf, "\talign-content: stretch;\n")
	}
	fmt.Fprintf(buf, "}\n")

	for i, m := range t.measured {
		fmt.Fprintf(buf, `#child%d {
	width: %.2fpx;
	height: %.2fpx;
`, i, m[0], m[1])
		c := colors[i%len(colors)]
		fmt.Fprintf(buf, "\tbackground-color: rgb(%d, %d, %d);\n", c.R, c.G, c.B)
		if t.layoutData != nil {
			d := t.layoutData[i]
			if d.MinSize.X != 0 {
				fmt.Fprintf(buf, "\tmin-width: %dpx;\n", d.MinSize.X)
			}
			if d.MinSize.Y != 0 {
				fmt.Fprintf(buf, "\tmin-height: %dpx;\n", d.MinSize.Y)
			}
			if d.MaxSize != nil {
				fmt.Fprintf(buf, "\tmax-width: %dpx;\n", d.MaxSize.X)
				fmt.Fprintf(buf, "\tmax-height: %dpx;\n", d.MaxSize.Y)
			}
			if d.Grow != 0 {
				fmt.Fprintf(buf, "\tflex-grow: %f;\n", d.Grow)
			}
			if d.Shrink != nil {
				fmt.Fprintf(buf, "\tflex-shrink: %f;\n", *d.Shrink)
			}
			// TODO: Basis, Align, BreakAfter
		}
		fmt.Fprintf(buf, "}\n")
	}
	fmt.Fprintf(buf, `</style>
<div id="container">
`)
	for i := range t.measured {
		fmt.Fprintf(buf, "\t<div id=\"child%d\"></div>\n", i)
	}
	fmt.Fprintf(buf, `</div>
<pre id="out"></pre>
<script>
var out = document.getElementById("out");
var container = document.getElementById("container");
for (var i = 0; i < container.children.length; i++) {
	var c = container.children[i];
	var ctop = c.offsetTop - container.offsetTop;
	var cleft = c.offsetLeft - container.offsetLeft;
	var cbottom = ctop + c.offsetHeight;
	var cright = cleft + c.offsetWidth;

	out.innerHTML += "\timage.Rect(" + cleft + ", " + ctop + ", " + cright + ", " + cbottom + "),\n";
}
</script>
`)

	return buf.String()
}

var colors = []color.RGBA{
	{0x00, 0x7f, 0x7f, 0xff}, // Cyan
	{0x7f, 0x00, 0x7f, 0xff}, // Magenta
	{0x7f, 0x7f, 0x00, 0xff}, // Yellow
	{0xff, 0x00, 0x00, 0xff}, // Red
	{0x00, 0xff, 0x00, 0xff}, // Green
	{0x00, 0x00, 0xff, 0xff}, // Blue
}

var layoutTests = []layoutTest{{
	size:     image.Point{100, 100},
	measured: [][2]float64{{100, 100}},
	want: []image.Rectangle{
		image.Rect(0, 0, 100, 100),
	},
}, {
	size:     image.Point{350, 100},
	measured: [][2]float64{{100, 100}, {100, 100}, {100, 100}},
	want: []image.Rectangle{
		image.Rect(0, 0, 100, 100),
		image.Rect(100, 0, 200, 100),
		image.Rect(200, 0, 300, 100),
	},
}, {
	size:     image.Point{300, 100},
	measured: [][2]float64{{100, 100}, {100, 100}},
	want: []image.Rectangle{
		image.Rect(0, 0, 100, 100),
		image.Rect(100, 0, 300, 100),
	},
	layoutData: []LayoutData{{}, {Grow: 1}},
}, {
	size:     image.Point{300, 100},
	measured: [][2]float64{{50, 50}, {100, 100}, {100, 100}},
	want: []image.Rectangle{
		image.Rect(0, 0, 50, 50),
		image.Rect(50, 0, 175, 100),
		image.Rect(175, 0, 300, 100),
	},
	layoutData: []LayoutData{{}, {Grow: 1}, {Grow: 1}},
}, {
	size:     image.Point{300, 100},
	measured: [][2]float64{{20, 100}, {20, 100}, {20, 100}},
	want: []image.Rectangle{
		image.Rect(0, 0, 30, 100),
		image.Rect(30, 0, 130, 100),
		image.Rect(130, 0, 300, 100),
	},
	layoutData: []LayoutData{
		{MaxSize: &image.Point{30, 100}, Grow: 1},
		{MinSize: image.Point{100, 0}, Grow: 1},
		{Grow: 4},
	},
}}

func TestLayout(t *testing.T) {
	for testNum, test := range layoutTests {
		w := NewFlex()
		w.Direction = test.direction
		w.Wrap = test.wrap
		w.AlignContent = test.alignContent

		var children []node.Node
		for i, sz := range test.measured {
			n := widget.NewUniform(colors[i], unit.Pixels(sz[0]), unit.Pixels(sz[1]))
			if test.layoutData != nil {
				n.LayoutData = test.layoutData[i]
			}
			w.AppendChild(n)
			children = append(children, n)
		}

		w.Measure(nil)
		w.Rect = image.Rectangle{Max: test.size}
		w.Layout(nil)

		bad := false
		for i, n := range children {
			if n.Wrappee().Rect != test.want[i] {
				bad = true
				break
			}
		}
		if bad {
			t.Logf("Bad testNum %d:\n%s", testNum, test.html())
		}
		for i, n := range children {
			if got, want := n.Wrappee().Rect, test.want[i]; got != want {
				t.Errorf("[%d].Rect=%v, want %v", i, got, want)
			}
		}
	}
}
