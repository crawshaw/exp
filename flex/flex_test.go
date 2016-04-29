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
			{size(0, 0), size(100, 100)},
			{size(100, 0), size(200, 100)},
			{size(200, 0), size(300, 100)},
		},
	},
	{
		size:     image.Point{300, 100},
		measured: [][2]float64{{100, 100}, {100, 100}},
		want: []image.Rectangle{
			{size(0, 0), size(100, 100)},
			{size(100, 0), size(300, 100)},
		},
		layoutData: []LayoutData{{}, {Grow: 1}},
	},
	{
		size:     image.Point{300, 100},
		measured: [][2]float64{{50, 50}, {100, 100}, {100, 100}},
		want: []image.Rectangle{
			{size(0, 0), size(50, 100)},
			{size(50, 0), size(175, 100)},
			{size(175, 0), size(300, 100)},
		},
		layoutData: []LayoutData{{}, {Grow: 1}, {Grow: 1}},
	},
	{
		size:     image.Point{300, 100},
		measured: [][2]float64{{20, 100}, {20, 100}, {20, 100}},
		want: []image.Rectangle{
			{size(0, 0), size(30, 100)},
			{size(30, 0), size(130, 100)},
			{size(130, 0), size(300, 100)},
		},
		layoutData: []LayoutData{
			{MaxSize: sizeptr(30, 100), Grow: 1},
			{MinSize: size(100, 0), Grow: 1},
			{Grow: 4}},
	},
}

func size(x, y int) image.Point { return image.Pt(x, y) }
func sizeptr(x, y int) *image.Point {
	s := size(x, y)
	return &s
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
			if test.layoutData != nil {
				n.LayoutData = test.layoutData[i]
			}
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
