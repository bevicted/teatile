package teatile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tile := New()
	assert.NotNil(t, tile)
}

func TestSetAndGetSubtiles(t *testing.T) {
	t.Parallel()

	const name = "name"

	tile := New()
	sub1 := tile.NewSubtile()
	sub2 := tile.NewNamedSubtile(name)

	for _, sl := range tile.subtiles {
		assert.Contains(t, []*Tile{sub1, sub2}, sl)
		assert.Same(t, tile, sl.parent)
	}
	assert.Len(t, tile.subtiles, 2)
	assert.Same(t, sub2, tile.GetSubtile(name))
}

func TestSetAndGetSize(t *testing.T) {
	t.Parallel()

	const (
		setW = 1
		setH = 2
	)

	t.Run("tile starts empty", func(t *testing.T) {
		t.Parallel()
		w, h := New().GetSize()
		assert.Empty(t, w)
		assert.Empty(t, h)
	})
	t.Run("tile correctly returns set size", func(t *testing.T) {
		t.Parallel()
		w, h := New().WithSize(setW, setH).GetSize()
		assert.Equal(t, setW, w)
		assert.Equal(t, setH, h)
	})
	t.Run("subtile fills parent space", func(t *testing.T) {
		t.Parallel()
		w, h := New().WithSize(setW, setH).NewSubtile().GetSize()
		assert.Equal(t, setW, w)
		assert.Equal(t, setH, h)
	})
	t.Run("subtile without set parent space cant calc what to fill", func(t *testing.T) {
		t.Parallel()
		w, h := New().NewSubtile().GetSize()
		assert.Empty(t, w)
		assert.Empty(t, h)
	})
}

func TestSizeCalculations(t *testing.T) {
	t.Parallel()

	parent := &Tile{setWidth: 10, setHeight: 15}

	for tcIdx, tc := range []struct {
		tiles     []*Tile
		expectedW []int
		expectedH []int
	}{
		{
			tiles:     []*Tile{{}, {}},
			expectedW: []int{5, 5},
			expectedH: []int{7, 8},
		},
		{
			tiles:     []*Tile{{}, {}, {}},
			expectedW: []int{3, 3, 4},
			expectedH: []int{5, 5, 5},
		},
		{
			tiles: []*Tile{
				{},
				{setWidth: 10, setHeight: 10},
				{setHeight: 3},
			},
			expectedW: []int{0, 10, 0},
			expectedH: []int{2, 10, 3},
		},
		{
			tiles: []*Tile{
				{setWidth: 5, setHeight: 5},
				{},
				{},
			},
			expectedW: []int{5, 2, 3},
			expectedH: []int{5, 5, 5},
		},
		{
			tiles: []*Tile{
				{},
				{},
				{},
				{},
				{},
				{setWidth: 5, setHeight: 10},
			},
			expectedW: []int{1, 1, 1, 1, 1, 5},
			expectedH: []int{1, 1, 1, 1, 1, 10},
		},
	} {
		t.Run(fmt.Sprintf("calc %d", tcIdx), func(t *testing.T) {
			tc := tc
			t.Parallel()

			// join tiles both vertically and horizontally
			// this doesn't really make sense for actual usage
			// but it's fine for testing
			for idx, tile := range tc.tiles {
				tile.parent = parent
				if idx != 0 {
					prev := tc.tiles[idx-1]
					prev.right = tile
					tile.left = prev
					prev.down = tile
					tile.up = prev
				}
			}

			require.Len(t, tc.expectedW, len(tc.tiles))
			require.Len(t, tc.expectedH, len(tc.tiles))

			for idx, tile := range tc.tiles {
				w, h := tile.GetSize()
				assert.Equal(t, tc.expectedW[idx], w, fmt.Sprintf("tcs[%d].expectedW[%d]", tcIdx, idx))
				assert.Equal(t, tc.expectedH[idx], h, fmt.Sprintf("tcs[%d].expectedH[%d]", tcIdx, idx))
			}
		})
	}
}

func TestRecalculate(t *testing.T) {
	t.Parallel()

	// setup
	tile := New().WithSize(10, 10)
	sub1 := tile.NewSubtile().WithSize(3, 0)
	sub2 := tile.NewSubtile()
	sub3 := tile.NewSubtile().WithSize(0, 3)
	cbCallCount := 0
	sub2.OnRecalculate(func() { cbCallCount++ })
	JoinHorizontal(sub1, sub2)
	JoinVertical(sub2, sub3)

	t.Run("sizes are calculated correctly", func(t *testing.T) {
		w1, h1 := sub1.GetSize()
		assert.Equal(t, 3, w1)
		assert.Equal(t, 10, h1)

		w2, h2 := sub2.GetSize()
		assert.Equal(t, 7, w2)
		assert.Equal(t, 7, h2)

		w3, h3 := sub3.GetSize()
		assert.Equal(t, 10, w3)
		assert.Equal(t, 3, h3)

		assert.Equal(t, 0, cbCallCount)
	})

	// sizes are re-set
	sub1.WithSize(0, 0)
	sub2.WithSize(8, 8)
	sub3.WithSize(0, 0)

	t.Run("sizes are incorrect before recalculate call", func(t *testing.T) {
		w1, h1 := sub1.GetSize()
		assert.Equal(t, 3, w1)
		assert.Equal(t, 10, h1)

		w2, h2 := sub2.GetSize()
		assert.Equal(t, 8, w2)
		assert.Equal(t, 8, h2)

		w3, h3 := sub3.GetSize()
		assert.Equal(t, 10, w3)
		assert.Equal(t, 3, h3)

		assert.Equal(t, 0, cbCallCount)
	})

	tile.Recalculate()

	t.Run("sizes are correct again after recalculate call", func(t *testing.T) {
		w1, h1 := sub1.GetSize()
		assert.Equal(t, 2, w1)
		assert.Equal(t, 10, h1)

		w2, h2 := sub2.GetSize()
		assert.Equal(t, 8, w2)
		assert.Equal(t, 8, h2)

		w3, h3 := sub3.GetSize()
		assert.Equal(t, 10, w3)
		assert.Equal(t, 2, h3)

		assert.Equal(t, 1, cbCallCount)
	})
}

func TestCalculations(t *testing.T) {
	t.Parallel()

	lyt := New().WithSize(120, 60)
	sub1 := lyt.NewSubtile()
	sub2 := lyt.NewNamedSubtile("sub2")
	JoinHorizontal(sub1, sub2)
	assert.Len(t, lyt.subtiles, 2)
	assert.Same(t, sub2, sub1.right)
	assert.Same(t, sub1, sub2.left)
	var idx int
	for l := range iterH(sub2) {
		assert.Same(t, lyt, l.parent)
		switch idx {
		case 0:
			assert.Same(t, sub1, l)
		case 1:
			assert.Same(t, sub2, l)
		default:
			assert.Fail(t, "more horizontal tiles than added")
		}
		idx++
	}
	JoinVertical(sub1, sub2)

	w, h := sub1.GetSize()
	assert.Equal(t, 60, w, "sub1 w")
	assert.Equal(t, 30, h, "sub1 h")
	w, h = sub2.GetSize()
	assert.Equal(t, 60, w, "sub2 w")
	assert.Equal(t, 30, h, "sub2 h")

	lyt.WithSize(300, 63).Recalculate()
	assert.Empty(t, sub1.setWidth)
	assert.Empty(t, lyt.GetSubtile("sub2").setWidth)

	w, h = sub1.GetSize()
	assert.Equal(t, 150, w, "sub1 recalc w")
	assert.Equal(t, 31, h, "sub1 recalc h")
	w, h = lyt.GetSubtile("sub2").GetSize()
	assert.Equal(t, 150, w, "sub2 recalc w")
	assert.Equal(t, 32, h, "sub2 recalc h")
}

func TestJoinAndIterH(t *testing.T) {
	t.Parallel()

	tile := []*Tile{}
	for range 10 {
		tile = append(tile, New())
	}
	JoinHorizontal(tile...)

	htile := []*Tile{}
	for tile := range iterH(tile[0]) {
		htile = append(htile, tile)
	}
	htile2 := []*Tile{}
	for tile := range iterH(tile[0]) {
		htile2 = append(htile2, tile)
		if len(htile2) == 3 {
			break
		}
	}

	assert.Equal(t, tile, htile)
	assert.Equal(t, tile[:3], htile2)
}

func TestJoinAndIterV(t *testing.T) {
	t.Parallel()

	tile := []*Tile{}
	for range 10 {
		tile = append(tile, New())
	}
	JoinVertical(tile...)

	vtile := []*Tile{}
	for tile := range iterV(tile[0]) {
		vtile = append(vtile, tile)
	}

	vtile2 := []*Tile{}
	for tile := range iterV(tile[0]) {
		vtile2 = append(vtile2, tile)
		if len(vtile2) == 3 {
			break
		}
	}

	assert.Equal(t, tile, vtile)
	assert.Equal(t, tile[:3], vtile2)
}

func TestScenario(t *testing.T) {
	t.Parallel()

	// ui
	lyt := New().WithSize(30, 60)
	headerTile := lyt.NewSubtile().WithSize(0, 2)
	contentTile := lyt.NewSubtile()
	footerTile := lyt.NewSubtile().WithSize(0, 1)
	JoinVertical(headerTile, contentTile, footerTile)

	// instances -> list
	listTile := contentTile.NewSubtile()
	fuzzyTile := listTile.NewSubtile().WithSize(0, 2)
	listItemsTile := listTile.NewSubtile()
	JoinVertical(fuzzyTile, listItemsTile)

	w, h := lyt.GetSize()
	assert.Equal(t, 30, w)
	assert.Equal(t, 60, h)

	w, h = headerTile.GetSize()
	assert.Equal(t, 30, w)
	assert.Equal(t, 2, h)

	w, h = footerTile.GetSize()
	assert.Equal(t, 30, w)
	assert.Equal(t, 1, h)

	w, h = contentTile.GetSize()
	assert.Equal(t, 30, w)
	assert.Equal(t, 57, h)

	w, h = listTile.GetSize()
	assert.Equal(t, 30, w)
	assert.Equal(t, 57, h)

	w, h = fuzzyTile.GetSize()
	assert.Equal(t, 30, w)
	assert.Equal(t, 2, h)

	w, h = listItemsTile.GetSize()
	assert.Equal(t, 30, w)
	assert.Equal(t, 55, h)
}
