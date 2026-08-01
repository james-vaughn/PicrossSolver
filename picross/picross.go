package picross

type Picross struct {
	Width     int
	Height    int
	Grid      [][]bool
	KeyAcross [][]int
	KeyDown   [][]int
}

func NewPicross(width, height int) *Picross {
	grid := make([][]bool, height)
	for i := range grid {
		grid[i] = make([]bool, width)
	}
	return &Picross{
		Width:     width,
		Height:    height,
		Grid:      grid,
		KeyAcross: make([][]int, width),
		KeyDown:   make([][]int, height),
	}
}

func (p Picross) String() string {
	result := ""
	for _, row := range p.Grid {
		for _, cell := range row {
			if cell {
				result += "X "
			} else {
				result += ". "
			}
		}
		result += "\n"
	}
	return result
}

func (p *Picross) Validate() bool {
	for i, row := range p.Grid {
		if len(row) != p.Width {
			return false
		}

		keyForRow := p.KeyDown[i]
		if !validateDim(row, keyForRow) {
			return false
		}
	}
	if len(p.Grid) != p.Height {
		return false
	}

	//validate columns
	for j := 0; j < p.Width; j++ {
		column := make([]bool, p.Height)
		for i := 0; i < p.Height; i++ {
			column[i] = p.Grid[i][j]
		}
		keyForColumn := p.KeyAcross[j]
		if !validateDim(column, keyForColumn) {
			return false
		}
	}

	return true
}

func validateDim(row []bool, key []int) bool {
	currentKey := key[0]
	count := 0

	for _, val := range row {
		if val {
			count++
		} else {
			if count != currentKey {
				return false
			}

			count = 0
		}
	}

	return true
}
