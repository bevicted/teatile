# TeaTile

![Latest Release](https://img.shields.io/github/v/release/bevicted/teatile)
[![GoDoc](https://pkg.go.dev/badge/github.com/bevicted/teatile.svg)](https://pkg.go.dev/github.com/bevicted/teatile)
[![Build](https://github.com/bevicted/teatile/actions/workflows/go.yml/badge.svg)](https://github.com/bevicted/teatile/actions/workflows/go.yml)

Tiling layout manager for
[Bubble Tea](https://github.com/charmbracelet/bubbletea).

If you're using multiple `tea.Model`s you might've run into the annoying
situation where you have to pass around `tea.WindowSizeMsg` literally
everywhere and keep track of what's taking how much space. Forgot to set width
somewhere? Sorry, you're TUI will now break in half, good luck finding the
problem!

With `teatile` you can assign a `Tile` to hold a rectangular space's width and
height and recalculate it if necessary.

## Why TeaTile?

TeaTile aims to be minimal and simplistic. It has no dependencies (besides
testify), provides a single struct with a constructor and 7 methods.
TeaTile also doesn't try to do anything fancy, it tries to fill a rectangular
space, that's it.

## Usage

```sh
go get -u github.com/bevicted/teatile@latest
```

[GoDoc](https://pkg.go.dev/github.com/bevicted/teatile)

### Basic

```go

// given that our main Tile has a height of 10 lines
tile := teatile.New().WithSize(0, 10)
subtile := tile.NewSubtile()

// we have 3 subtiles joined vertically
teatile.JoinVertical(
	// takes 3 lines of space
	tile.NewSubtile().WithSize(0, 3),

	// with no set height, these Tiles will fill the remaining space

	// fills (10 - 3) / 2 = 3 lines
	subtile,
	// fills the remaining 10 - 3 - 3 = 4 lines
	tile.NewSubtile(),
)
```

### Example

In your main `tea.Model`, instead of holding the width and height as a field, hold your main `Tile`.

```go
// I prefer saving multiple Tiles into a single struct, but you do you
type tiles struct {
	main   *teatile.Tile
	header *teatile.Tile
}

type Model struct {
	tiles tiles
...

func New() Model {
	// create your main Tile
	tile := teatile.New()

    // create a subtiles, set sizes where necessary
	// here we skip setting width with 0 and set height to 1 line
	headerTile := tile.NewSubtile().WithSize(0, 1)
	contentTile := tile.NewSubtile()
	footerTile := tile.NewSubtile().WithSize(0, 1)

	// tell the tiles that they are vertically next to eachother
	teatile.JoinVertical(headerTile, contentTile, footerTile)

	m := &Model{
		// save tiles we must reference later
		tiles: tiles{
			main:   tile,
			header: headerTile,
		},
		// pass down the tiles submodels need to fill
		contentModel: contentModel.New(contentTile),
		footerModel:  footer.New(footerTile),
...

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// on window size change, set the size of the main Tile and tell it
		// to recalculate the occupied area. This will propagate to all
		// linked Tiles, be it horizontally/vertically joined or subtiles.
		m.tiles.main.WithSize(msg.Width, msg.Height).Recalculate()
...

// use OnRecalculate to connect custom components
func (m *Model) Init() tea.Cmd {
	m.tiles.main.OnRecalculate(func() {
		w, h := m.tiles.main.GetSize()
		m.textArea.SetWidth(w)
		m.textArea.SetHeight(h)
	})
	return nil
}

```

