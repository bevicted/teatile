package teatile

import (
	"fmt"

	"iter"

	"github.com/google/uuid"
)

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
	subtiles  map[string]*Tile
	recalcCBs []func()
}

// New creates a new Tile.
func New() *Tile {
	return &Tile{
		subtiles: map[string]*Tile{},
	}
}

// WithSize sets the width and height of the tile.
func (t *Tile) WithSize(w, h int) *Tile {
	t.setWidth = w
	t.setHeight = h
	return t
}

// getSize returns the width and height of the tile
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

// GetSize returns the width and height of the tile and calculates them if necessary.
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

// NewNamedSubtile creates a new subtile with a name that can be used to refer to it later.
func (t *Tile) NewNamedSubtile(name string) *Tile {
	subtile := New()
	t.subtiles[name] = subtile
	subtile.parent = t
	return subtile
}

// NewSubtile creates a new subtile with a random unique name, use this if there's no need
// to refer to the tile later.
func (t *Tile) NewSubtile() *Tile {
	return t.NewNamedSubtile(uuid.NewString())
}

// GetSubtile returns the subtile by name.
func (t *Tile) GetSubtile(name string) *Tile {
	sl, ok := t.subtiles[name]
	if !ok {
		panic(fmt.Sprintf("no such subtile: %q", name))
	}
	return sl
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
// the horizontal chain.
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
