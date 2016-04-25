// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flex

import (
	"fmt"
	"image"

	"golang.org/x/exp/shiny/widget"
)

// Flex is a container widget that lays out its children using the
// CSS flexbox algorithm.
type Flex struct {
	widget.Node

	Direction    ContainerDirection
	Wrap         ContainerWrap
	Justify      ContainerJustify
	AlignItem    AlignItem
	AlignContent ContainerAlignContent
}

// NewFlex returns a new Flex widget.
func NewFlex() *Flex {
	fl := new(Flex)
	fl.Node.Class = &flexClass{flex: fl}
	return fl
}

// ContainerDirection
//
// https://www.w3.org/TR/css-flexbox-1/#flex-direction-property
type ContainerDirection int8

const (
	Row ContainerDirection = iota
	RowReverse
	Column
	ColumnReverse
)

// ContainerWrap
//
// https://www.w3.org/TR/css-flexbox-1/#flex-wrap-property
type ContainerWrap int8

const (
	NoWrap ContainerWrap = iota
	Wrap
	WrapReverse
)

// ContainerJustify
//
// https://www.w3.org/TR/css-flexbox-1/#justify-content-property
type ContainerJustify int8

const (
	JustifyStart ContainerJustify = iota
	JustifyEnd
	JustifyCenter
	JustifySpaceBetween
	JustifySpaceAround
)

// AlignItem
//
// https://www.w3.org/TR/css-flexbox-1/#align-items-property
type AlignItem int8

const (
	AlignItemAuto AlignItem = iota
	AlignItemStart
	AlignItemEnd
	AlignItemCenter
	AlignItemBaseline
	AlignItemStretch
)

// ContainerAlignContent
//
// https://www.w3.org/TR/css-flexbox-1/#align-content-property
type ContainerAlignContent int8

const (
	AlignContentStart ContainerAlignContent = iota
	AlignContentEnd
	AlignContentCenter
	AlignContentSpaceBetween
	AlignContentSpaceAround
	AlignContentStretch
)

type flexClass struct {
	widget.ContainerClassEmbed

	flex *Flex
}

func (k *flexClass) Measure(n *widget.Node, t *widget.Theme) {
	// As Measure is a bottom-up calculation of natural size, we have no
	// hint yet as to how we should flex. So we ignore Wrap, Justify,
	// AlignItem, AlignContent.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if d, ok := c.LayoutData.(FlexLayoutData); ok {
			_ = d
			panic("TODO Measure")
		}
	}
}

func (k *flexClass) Layout(n *widget.Node, t *widget.Theme) {
	type element struct {
		n *widget.Node
		flexBaseSize int
		frozen bool
		mainSize int
		crossSize int
	}
	var children []element

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		var flexBaseSize int // §9.2.3
		basis := Auto
		if d, ok := c.LayoutData.(FlexLayoutData); ok {
			basis = d.Basis
		}
		switch basis {
		case Definite: // A
			flexBaseSize = c.LayoutData.(FlexLayoutData).BasisPx // A
		case Content:
			// TODO §9.2.3.B
			// TODO §9.2.3.C
			// TODO §9.2.3.D
			panic("flex-basis: content not supported")
		case Auto: // E
			flexBaseSize = k.mainSize(c.MeasuredSize)
		}
		children = append(children, element{
			flexBaseSize: flexBaseSize,
			n: c,
		})
	}

	flexBaseSize := k.mainSize(n.Rect.Size())
	hypotheticalMainSize := flexBaseSize // no min/max properties to clamp

	// §9.3.5 collect children into flex lines
	type flexLine struct {
		mainSize   int // TODO move out?
		child      []*element
	}
	var lines []flexLine
	if k.flex.Wrap == NoWrap {
		line := flexLine{ child: make([]*element, len(children)) }
		for i := range children {
			line.child[i] = &children[i]
		}
		lines = []flexLine{line}
	} else {
		var line flexLine

		for i := range children {
			child := &children[i]
			if line.mainSize > 0 && line.mainSize+child.flexBaseSize > hypotheticalMainSize {
				lines = append(lines, line)
				line = flexLine{}
			}
			line.child = append(line.child, child)
			line.mainSize += child.flexBaseSize

			if d, ok := child.n.LayoutData.(FlexLayoutData); ok && d.BreakAfter {
				lines = append(lines, line)
				line = flexLine{}
			}
		}

		if k.flex.Wrap == WrapReverse {
			for i := 0; i < len(lines)/2; i++ {
				lines[i], lines[len(lines)-i-1] = lines[len(lines)-i-1], lines[i]
			}
		}
	}

	// §9.3.6 resolve flexible lengths (details in section §9.7)
	for lineNum := range lines {
		line := &lines[lineNum]
		grow := line.mainSize < hypotheticalMainSize // §9.7.1

		// §9.7.2 freeze inflexible children.
		for _, child := range line.child {
			mainSize := k.mainSize(child.n.MeasuredSize)
			if grow {
				if growFactor(child.n) == 0 || baseSize(child.n) > mainSize {
					child.frozen = true
					child.mainSize = mainSize
				}
			} else {
				if shrinkFactor(child.n) == 0 || baseSize(child.n) < mainSize {
					child.frozen = true
					child.mainSize = mainSize
				}
			}
		}
	}
}

// TODO methods on element?

func baseSize(n *widget.Node) int {
	panic("TODO")
}

func growFactor(n *widget.Node) int {
	if d, ok := n.LayoutData.(FlexLayoutData); ok {
		return d.Grow
	}
	return 0
}

func shrinkFactor(n *widget.Node) int {
	if d, ok := n.LayoutData.(FlexLayoutData); ok && d.Shrink != nil {
		return *d.Shrink
	}
	return 1
}

func (k *flexClass) mainSize(p image.Point) int {
	switch k.flex.Direction {
	case Row, RowReverse:
		return p.X
	case Column, ColumnReverse:
		return p.Y
	default:
		panic(fmt.Sprint("bad direction: ", k.flex.Direction))
	}
}

type Basis int8

const (
	Auto    Basis = iota
	Content       // TODO
	Definite
)

// FlexLayoutData is the Node.LayoutData type for a Flex's children.
type FlexLayoutData struct {
	// Grow is the flex grow factor which determines how much a Node
	// will grow relative to its siblings.
	Grow int

	// Shrink is the flex shrink factor which determines how much a Node
	// will shrink relative to its siblings. If nil, a default shrink
	// factor of 1 is used.
	Shrink *int

	// Basis determines the initial main size of the of the Node.
	// If set to Definite, the value stored in BasisPx is used.
	Basis   Basis
	BasisPx int

	Align AlignItem

	// BreakAfter forces the next node onto the next flex line.
	BreakAfter bool
}
