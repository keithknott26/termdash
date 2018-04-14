// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package draw

// box.go contains code that draws boxes.

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
)

// boxChar returns the correct box character from the parts for the use at the
// specified point of the box. Returns -1 if no character should be at this point.
func boxChar(p image.Point, box image.Rectangle, parts map[linePart]rune) rune {
	switch {
	case p.X == box.Min.X && p.Y == box.Min.Y:
		return parts[topLeftCorner]
	case p.X == box.Max.X-1 && p.Y == box.Min.Y:
		return parts[topRightCorner]
	case p.X == box.Min.X && p.Y == box.Max.Y-1:
		return parts[bottomLeftCorner]
	case p.X == box.Max.X-1 && p.Y == box.Max.Y-1:
		return parts[bottomRightCorner]
	case p.X == box.Min.X || p.X == box.Max.X-1:
		return parts[vLine]
	case p.Y == box.Min.Y || p.Y == box.Max.Y-1:
		return parts[hLine]
	}
	return -1
}

// Box draws a box on the canvas.
func Box(c *canvas.Canvas, box image.Rectangle, ls LineStyle, opts ...cell.Option) error {
	if ar := c.Area(); !box.In(ar) {
		return fmt.Errorf("the requested box %+v falls outside of the provided canvas %+v", box, ar)
	}

	const minSize = 2
	if box.Dx() < minSize || box.Dy() < minSize {
		return fmt.Errorf("the smallest supported box is %dx%d, got: %dx%d", minSize, minSize, box.Dx(), box.Dy())
	}

	parts, err := lineParts(ls)
	if err != nil {
		return err
	}

	for col := box.Min.X; col < box.Max.X; col++ {
		for row := box.Min.Y; row < box.Max.Y; row++ {
			p := image.Point{col, row}
			r := boxChar(p, box, parts)
			if r == -1 {
				continue
			}

			if err := c.SetCell(p, r, opts...); err != nil {
				return err
			}
		}
	}
	return nil
}