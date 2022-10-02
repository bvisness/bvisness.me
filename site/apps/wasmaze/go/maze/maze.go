package maze

import "math/rand"

type Coords struct {
	X, Y int
}

type Cell struct {
	ConnectsTo []Coords
}

func NewCell() Cell {
	return Cell{
		ConnectsTo: make([]Coords, 0, 4),
	}
}

type Maze struct {
	Width, Height int
	Cells         [][]Cell
}

func GenMaze(width, height int) Maze {
	cells := make([][]Cell, width)
	for x := 0; x < width; x++ {
		cells[x] = make([]Cell, height)

		for y := 0; y < height; y++ {
			cells[x][y] = NewCell()
		}
	}

	visited := make([][]bool, width)
	for x := 0; x < width; x++ {
		visited[x] = make([]bool, height)
	}

	start := Coords{rand.Intn(width), rand.Intn(height)}
	stack := []Coords{start}

	unvisitedNeighbors := make([]Coords, 0, 4)

	for len(stack) > 0 {
		c := stack[len(stack)-1]

		visited[c.X][c.Y] = true

		unvisitedNeighbors = unvisitedNeighbors[:0]
		for _, d := range []Coords{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			newX := c.X + d.X
			newY := c.Y + d.Y

			if newX < 0 || newX >= width || newY < 0 || newY >= height || visited[newX][newY] {
				continue
			} else {
				unvisitedNeighbors = append(unvisitedNeighbors, Coords{newX, newY})
			}
		}

		if len(unvisitedNeighbors) > 0 {
			n := unvisitedNeighbors[rand.Intn(len(unvisitedNeighbors))]

			currentCell := &cells[c.X][c.Y]
			currentCell.ConnectsTo = append(currentCell.ConnectsTo, n)

			neighborCell := &cells[n.X][n.Y]
			neighborCell.ConnectsTo = append(neighborCell.ConnectsTo, c)

			stack = append(stack, n)
		} else {
			stack = stack[:len(stack)-1]
		}
	}

	return Maze{
		Width:  width,
		Height: height,
		Cells:  cells,
	}
}
