// Package teatile provides a simple layout management system for Bubble Tea
// applications.
//
// It allows allocating and managing rectangular spaces (tiles) that can be nested and sized
// automatically, which is useful for building complex TUI layouts with Bubble Tea.
package teatile

import (
	"iter"
)

type style[S any] interface {
	Width(int) S
	MaxWidth(int) S
	Height(int) S
	MaxHeight(int) S
}

// SetStyleWidth sets the Width and MaxWidth of style to the width of Tile.
func SetStyleWidth[S style[S]](s S, t *Tile) S {
	w, _ := t.GetSize()
	return s.Width(w).MaxWidth(w)
}

// SetStyleHeight sets the Height and MaxHeight of style to the height of Tile.
func SetStyleHeight[S style[S]](s S, t *Tile) S {
	_, h := t.GetSize()
	return s.Height(h).MaxHeight(h)
}

// SetStyleSize sets the Width, MaxWidth, Height and MaxHeight of style to
// the Tile's size.
func SetStyleSize[S style[S]](s S, t *Tile) S {
	w, h := t.GetSize()
	return s.Width(w).MaxWidth(w).Height(h).MaxHeight(h)
}

// Tile represents a rectangular space in the layout.
// Subtiles will strive to fill their parent's space.
type Tile struct {
	setWidth  int
	setHeight int
	calcW     int
	calcH     int

	parent    *Tile
	up        *Tile
	right     *Tile
	down      *Tile
	left      *Tile
	subtiles  []*Tile
	recalcCBs []func()
}

// New creates a new Tile and returns a pointer to it.
func New() *Tile {
	return &Tile{}
}

// WithSize sets the width and height for the Tile.
func (t *Tile) WithSize(w, h int) *Tile {
	t.setWidth = w
	t.setHeight = h
	return t
}

// getSize returns the width and height of the Tile.
//
// It first considers explicitly set sizes and if those are zero, returns the calculated sizes.
func (t *Tile) getSize() (int, int) {
	w := t.setWidth
	if w == 0 {
		w = t.calcW
	}

	h := t.setHeight
	if h == 0 {
		h = t.calcH
	}

	return w, h
}

// GetSize returns the width and height of the Tile, calculating them if necessary.
func (t *Tile) GetSize() (int, int) {
	w, h := t.getSize()

	if w != 0 && h != 0 {
		// nothing to calculate
		return w, h
	}

	if t.parent == nil {
		// can't calculate size, since we don't know the dimensions to fill
		return w, h
	}

	parentW, parentH := t.parent.GetSize()
	if parentW == 0 || parentH == 0 {
		// can't calculate size, since parent has no size
		// this might happen before the first window size msg arrives
		return 0, 0
	}

	if w == 0 {
		var (
			unsetWCount    int
			allocatedSpace int
		)
		for sl := range iterH(t) {
			sw, _ := sl.getSize()
			if sw != 0 {
				allocatedSpace += sw
			} else {
				unsetWCount++
			}
		}

		spaceToAllocate := parentW - allocatedSpace
		switch unsetWCount {
		case 1:
			// space to allocate is right as-is
			w = spaceToAllocate
		case 2:
			// performant /2
			w = spaceToAllocate >> 1
		default:
			w = spaceToAllocate / unsetWCount
		}
	}

	if h == 0 {
		var (
			unsetHCount    int
			allocatedSpace int
		)
		for sl := range iterV(t) {
			_, sh := sl.getSize()
			if sh != 0 {
				allocatedSpace += sh
			} else {
				unsetHCount++
			}
		}

		spaceToAllocate := parentH - allocatedSpace
		switch unsetHCount {
		case 1:
			// space to allocate is right as-is
			h = spaceToAllocate
		case 2:
			// performant /2
			h = spaceToAllocate >> 1
		default:
			h = spaceToAllocate / unsetHCount
		}
	}

	t.calcW = w
	t.calcH = h

	return w, h
}

// NewSubtile creates a new subtile.
func (t *Tile) NewSubtile() *Tile {
	st := New()
	t.subtiles = append(t.subtiles, st)
	st.parent = t
	return st
}

// OnRecalculate adds the specified callback function to the list of funcs that will be called
// on recalculation of the tile's size.
func (t *Tile) OnRecalculate(cb func()) {
	t.recalcCBs = append(t.recalcCBs, cb)
}

// Recalculate forces a recalculation of the tile's size and calls OnRecalculate callbacks.
// This effect is recursively propagated to all subtiles as well.
func (t *Tile) Recalculate() {
	t.calcW = 0
	t.calcH = 0
	for _, sl := range t.subtiles {
		sl.Recalculate()
	}
	for _, cb := range t.recalcCBs {
		cb()
	}
}

// JoinHorizontal joins the specified tiles horizontally.
func JoinHorizontal(tiles ...*Tile) {
	for idx, tile := range tiles {
		if idx != 0 {
			prev := tiles[idx-1]
			prev.right = tile
			tile.left = prev
		}
	}
}

// JoinVertical joins the specified tiles vertically.
func JoinVertical(tiles ...*Tile) {
	for idx, lo := range tiles {
		if idx != 0 {
			prev := tiles[idx-1]
			prev.down = lo
			lo.up = prev
		}
	}
}

// iterH creates an iterator that starts from the lefmost joined tile
// and goes over every tile that's joined to the right of the current one.
//
// NOTE: tile given as parameter is not the starting point, simply a link in
// the horizontal chain.
func iterH(sibling *Tile) iter.Seq[*Tile] {
	tile := sibling
	for tile.left != nil {
		tile = tile.left
	}

	return func(yield func(*Tile) bool) {
		for tile != nil {
			if !yield(tile) {
				return
			}
			tile = tile.right
		}
	}
}

// iterV creates an iterator that starts from the upmost joined tile
// and goes over every tile that's joined to the bottom of the current one.
//
// NOTE: tile given as parameter is not the starting point, simply a link in
// the vertical chain.
func iterV(sibling *Tile) iter.Seq[*Tile] {
	tile := sibling
	for tile.up != nil {
		tile = tile.up
	}

	return func(yield func(*Tile) bool) {
		for tile != nil {
			if !yield(tile) {
				return
			}
			tile = tile.down
		}
	}
}
