// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package widget provides graphical user interface widgets.
//
// TODO: give an overview and some example code.
package widget // import "golang.org/x/exp/shiny/widget"

import (
	"image"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// Axis is zero, one or both of the horizontal and vertical axes. For example,
// a widget may be scrollable in one of the four AxisXxx values.
type Axis uint8

const (
	AxisNone       = Axis(0)
	AxisHorizontal = Axis(1)
	AxisVertical   = Axis(2)
	AxisBoth       = Axis(3) // AxisBoth equals AxisHorizontal | AxisVertical.
)

// TODO: how does RunWindow's caller inject or process events (whether general
// like lifecycle events or app-specific)? How does it stop the event loop when
// the app's work is done?

// TODO: how do widgets signal that they need repaint or relayout?

// TODO: propagate keyboard / mouse / touch events.

// RunWindow creates a new window for s, with the given widget tree, and runs
// its event loop.
func RunWindow(s screen.Screen, root node.Node) error {
	var (
		buf screen.Buffer
		t   theme.Theme
	)
	defer func() {
		if buf != nil {
			buf.Release()
		}
	}()

	w, err := s.NewWindow(nil)
	if err != nil {
		return err
	}
	defer w.Release()
	for {
		switch e := w.NextEvent().(type) {
		case lifecycle.Event:
			if e.To == lifecycle.StageDead {
				return nil
			}

		case mouse.Event:
			root.OnMouseEvent(e, image.Point{})

		case paint.Event:
			if buf != nil {
				w.Upload(image.Point{}, buf, buf.Bounds())
			}
			w.Publish()

		case size.Event:
			if buf != nil {
				buf.Release()
			}
			var err error
			buf, err = s.NewBuffer(e.Size())
			if err != nil {
				return err
			}
			t.DPI = float64(e.PixelsPerPt) * unit.PointsPerInch
			root.Measure(&t)
			root.Wrappee().Rect = e.Bounds()
			root.Layout(&t)
			root.Paint(&t, buf.RGBA(), image.Point{})

		case error:
			return e
		}
	}
}
