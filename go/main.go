package main

import (
	"syscall/js"

	"./maze"
)

func CoordsToJS(c *maze.Coords) js.Value {
	return js.ValueOf([]interface{}{c.X, c.Y})
}

func CellToJS(c *maze.Cell) js.Value {
	var jsCt []interface{}
	for _, ct := range c.ConnectsTo {
		jsCt = append(jsCt, CoordsToJS(&ct))
	}

	return js.ValueOf(map[string]interface{}{
		"connectsTo": jsCt,
	})
}

func MazeToJS(m *maze.Maze) js.Value {
	cells := make([]interface{}, m.Width)
	for x := 0; x < m.Width; x++ {
		col := make([]interface{}, m.Height)

		for y := 0; y < m.Height; y++ {
			col[y] = CellToJS(&m.Cells[x][y])
		}

		cells[x] = col
	}

	result := js.ValueOf(map[string]interface{}{
		"width":  m.Width,
		"height": m.Height,
		"cells":  cells,
	})

	return result
}

func genMazeGo(args []js.Value) {
	width := args[0].Int()
	height := args[1].Int()
	callback := args[2]

	maze := maze.GenMaze(width, height)

	callback.Invoke(MazeToJS(&maze))
}

func genMazeGoSilent(args []js.Value) {
	width := args[0].Int()
	height := args[1].Int()
	callback := args[2]

	_ = maze.GenMaze(width, height)

	callback.Invoke(js.ValueOf("done"))
}

func registerCallbacks() {
	js.Global().Set("genMazeGo", js.NewCallback(genMazeGo))
	js.Global().Set("genMazeGoSilent", js.NewCallback(genMazeGoSilent))
}

func main() {
	c := make(chan struct{}, 0)

	println("WASM Go Initialized")
	// register functions
	registerCallbacks()
	<-c
}
