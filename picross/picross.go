package picross

type Picross struct {
	Width  int
	Height int
	Grid   [][]bool
}

func NewPicross(width, height int) *Picross {
	grid := make([][]bool, height)
	for i := range grid {
		grid[i] = make([]bool, width)
	}
	return &Picross{
		Width:  width,
		Height: height,
		Grid:   grid,
	}
}
