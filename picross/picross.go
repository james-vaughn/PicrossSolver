package picross

import (
	"fmt"
	"math/rand"
)

type PicrossGrid = [][]bool
type PicrossKey = []int

type Picross struct {
	Width     int
	Height    int
	Grid      PicrossGrid
	KeyAcross []PicrossKey
	KeyDown   []PicrossKey
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

func FromGrid(grid PicrossGrid) *Picross {
	width := len(grid[0])
	height := len(grid)

	picross := &Picross{
		Width:     width,
		Height:    height,
		Grid:      grid,
		KeyAcross: make([][]int, width),
		KeyDown:   make([][]int, height),
	}

	generateKeys(picross)
	return picross
}

func (p Picross) String() string {
	result := ""

	result += "Keys Across: "
	for _, key := range p.KeyAcross {
		result += fmt.Sprint(key) + " "
	}
	result += "\n"

	result += "Keys Down: "
	for _, key := range p.KeyDown {
		result += fmt.Sprint(key) + " "
	}
	result += "\n"

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
	var groups []int
	count := 0

	for _, val := range row {
		if val {
			count++
		} else if count > 0 {
			groups = append(groups, count)
			count = 0
		}
	}
	if count > 0 {
		groups = append(groups, count)
	}

	// A key of [0] means "no filled cells expected"
	if len(key) == 1 && key[0] == 0 {
		return len(groups) == 0
	}

	if len(groups) != len(key) {
		return false
	}
	for i, g := range groups {
		if g != key[i] {
			return false
		}
	}

	return true
}

func RandomPicross(width, height int) *Picross {
	p := NewPicross(width, height)
	// Fill the grid with random true/false values
	for i := range height {
		for j := range width {
			p.Grid[i][j] = rand.Intn(2) == 1
		}
	}

	generateKeys(p)
	return p
}
func generateKeys(p *Picross) {
	for i := 0; i < p.Height; i++ {
		p.KeyDown[i] = generateKey(p.Grid[i])
	}
	for j := 0; j < p.Width; j++ {
		column := make([]bool, p.Height)
		for i := 0; i < p.Height; i++ {
			column[i] = p.Grid[i][j]
		}
		p.KeyAcross[j] = generateKey(column)
	}
}

func generateKey(row []bool) []int {
	var key []int
	count := 0
	for _, val := range row {
		if val {
			count++
		} else if count > 0 {
			key = append(key, count)
			count = 0
		}
	}
	if count > 0 {
		key = append(key, count)
	}
	if len(key) == 0 {
		key = []int{0}
	}
	return key
}
