package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/james-vaughn/PicrossSolver/picross"
	"github.com/james-vaughn/PicrossSolver/solver"
)

func main() {
	// puzzle, err := LoadPicross("puzzle.txt")
	// if err != nil {
	// 	fmt.Println("Error loading puzzle:", err)
	// 	return
	// }

	puzzle := picross.RandomPicross(15, 15)
	if err := WritePicross("puzzleOut.txt", puzzle); err != nil {
		fmt.Printf("Error writing puzzle to file: %v", err)
	}

	fmt.Println("Starting solve...")
	picross, solved := solver.SolvePicross(puzzle, solver.SolverConfig{
		GenerationSize:  200,
		GenerationCount: 20000,
		MutationRate:    .075,
		ElitismCount:    4,
		TournamentSize:  3,
	})

	if solved {
		fmt.Println("Solved Picross puzzle:")
		fmt.Println(picross)
	} else {
		fmt.Println("Failed to solve Picross puzzle.")
		fmt.Println("Best attempt:")
		fmt.Println(picross)
	}
}

var groupRe = regexp.MustCompile(`\[([^\]]*)\]`)

// parseLine turns something like "[1, 2], [1, 2], [0], [4]" into
// [][]int{{1,2}, {1,2}, {0}, {4}}
func parseLine(line string) [][]int {
	matches := groupRe.FindAllStringSubmatch(line, -1)
	result := make([][]int, len(matches))

	for i, m := range matches {
		parts := strings.Split(m[1], ",")
		nums := make([]int, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			n, err := strconv.Atoi(p)
			if err != nil {
				continue // or return an error if you want strict parsing
			}
			nums = append(nums, n)
		}
		result[i] = nums
	}

	return result
}

// LoadPicross reads a file where line 1 is the "across" keys and
// line 2 is the "down" keys, and builds a Picross from it.
func LoadPicross(filename string) (*picross.Picross, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	if len(lines) < 2 {
		return nil, fmt.Errorf("expected 2 lines, got %d", len(lines))
	}

	across := parseLine(lines[0])
	down := parseLine(lines[1])

	width := len(across)
	height := len(down)

	p := picross.NewPicross(width, height)
	p.KeyAcross = across
	p.KeyDown = down

	return p, nil
}

func WritePicross(filename string, p *picross.Picross) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	_, err = writer.WriteString(p.String())
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}
	return nil
}
