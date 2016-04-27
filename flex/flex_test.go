// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flex

import (
	"image"
	"image/color"
	"testing"

	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
)

type layoutTest struct {
	direction  ContainerDirection
	wrap       ContainerWrap
	size       image.Point       // size of container
	measured   [][2]float64      // MeasuredSize of child elements
	layoutData []LayoutData      // LayoutData of child elements
	want       []image.Rectangle // final Rect of child elements
}

var tileColors = []color.RGBA{
	color.RGBA{0x00, 0x7f, 0x7f, 0xff}, // Cyan
	color.RGBA{0x7f, 0x00, 0x7f, 0xff}, // Magenta
	color.RGBA{0x7f, 0x7f, 0x00, 0xff}, // Yellow
	color.RGBA{0xff, 0x00, 0x00, 0xff},
	color.RGBA{0x00, 0xff, 0x00, 0xff},
	color.RGBA{0x00, 0x00, 0xff, 0xff},
}

var layoutTests = []layoutTest{
	{
		size:     image.Point{350, 100},
		measured: [][2]float64{{100, 100}, {100, 100}, {100, 100}},
		want: []image.Rectangle{
			{image.Pt(0, 0), image.Pt(100, 100)},
			{image.Pt(100, 0), image.Pt(200, 100)},
			{image.Pt(200, 0), image.Pt(300, 100)},
		},
	},
}

func TestLayout(t *testing.T) {
	for testNum, test := range layoutTests {
		t.Logf("Layout testNum %d", testNum)

		fl := NewFlex()
		fl.Direction = test.direction
		fl.Wrap = test.wrap

		var children []*widget.Node
		for i, sz := range test.measured {
			n := widget.NewUniform(tileColors[i], unit.Pixels(sz[0]), unit.Pixels(sz[1])).Node
			fl.AppendChild(n)
			children = append(children, n)
		}

		fl.Node.Class.Measure(&fl.Node, nil)
		fl.Node.Rect = image.Rectangle{Max: test.size}
		fl.Node.Class.Layout(&fl.Node, nil)

		for i, n := range children {
			if n.Rect != test.want[i] {
				t.Errorf("\tchildren[%d].Rect=%v, want %v", i, n.Rect, test.want[i])
			}
		}
	}
}
