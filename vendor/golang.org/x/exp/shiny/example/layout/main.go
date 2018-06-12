// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build example
//
// This build tag means that "go install golang.org/x/exp/shiny/..." doesn't
// install this example program. Use "go run main.go" to run it or "go install
// -tags=example" to install it.

// layout is an example of a laying out a widget node tree. Real GUI programs
// won't need to do this explicitly, as the shiny/widget package will
// coordinate with the shiny/screen package to call Measure, Layout and Paint
// as necessary, and will re-layout widgets when windows are re-sized. This
// program merely demonstrates how a widget node tree can be rendered onto a
// statically sized RGBA image, for visual verification of widget code without
// having to bring up and manually inspect an interactive GUI window.
package layout

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/theme"
)

var px = unit.Pixels

func main() {
	t := theme.Default

	// Make the widget node tree.
	hf := widget.NewFlow(widget.AxisHorizontal)
	hf.AppendChild(widget.NewLabel("Cyan:"))
	hf.AppendChild(widget.NewUniform(color.RGBA{0x00, 0x7f, 0x7f, 0xff}, px(0), px(20)))
	hf.LastChild.LayoutData = widget.FlowLayoutData{ExpandAlongWeight: 1}
	hf.AppendChild(widget.NewLabel("Magenta:"))
	hf.AppendChild(widget.NewUniform(color.RGBA{0x7f, 0x00, 0x7f, 0xff}, px(0), px(30)))
	hf.LastChild.LayoutData = widget.FlowLayoutData{ExpandAlongWeight: 2}
	hf.AppendChild(widget.NewLabel("Yellow:"))
	hf.AppendChild(widget.NewUniform(color.RGBA{0x7f, 0x7f, 0x00, 0xff}, px(0), px(40)))
	hf.LastChild.LayoutData = widget.FlowLayoutData{ExpandAlongWeight: 3}

	vf := widget.NewFlow(widget.AxisVertical)
	vf.AppendChild(widget.NewUniform(color.RGBA{0xff, 0x00, 0x00, 0xff}, px(80), px(40)))
	vf.AppendChild(widget.NewUniform(color.RGBA{0x00, 0xff, 0x00, 0xff}, px(50), px(50)))
	vf.AppendChild(widget.NewUniform(color.RGBA{0x00, 0x00, 0xff, 0xff}, px(20), px(60)))
	vf.AppendChild(hf)
	vf.LastChild.LayoutData = widget.FlowLayoutData{ExpandAcross: true}
	vf.AppendChild(widget.NewLabel(fmt.Sprintf(
		"The black rectangle is 1.5 inches x 1 inch when viewed at %v DPI.", t.GetDPI())))
	vf.AppendChild(widget.NewUniform(color.Black, unit.Inches(1.5), unit.Inches(1)))

	// Make the RGBA image.
	rgba := image.NewRGBA(image.Rect(0, 0, 640, 480))
	draw.Draw(rgba, rgba.Bounds(), t.GetPalette().Neutral, image.Point{}, draw.Src)

	// Measure, layout and paint.
	vf.Measure(t)
	vf.Rect = rgba.Bounds()
	vf.Layout(t)
	vf.Paint(t, rgba, image.Point{})

	// Encode to PNG.
	out, err := os.Create("out.png")
	if err != nil {
		log.Fatalf("os.Create: %v", err)
	}
	defer out.Close()
	if err := png.Encode(out, rgba); err != nil {
		log.Fatalf("png.Encode: %v", err)
	}
	fmt.Println("Wrote out.png OK.")
}
