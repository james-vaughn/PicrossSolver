package solver

import (
	"fmt"
	"sort"
	"sync"

	"github.com/james-vaughn/PicrossSolver/picross"
)

type SolverConfig struct {
	GenerationSize  int
	GenerationCount int
	MutationRate    float64
}

type Individual struct {
	Genome  *picross.Picross
	Fitness int
}

func SolvePicross(p *picross.Picross, config SolverConfig) {
	// Implement the genetic algorithm to solve the Picross puzzle
	// This is a placeholder for the actual implementation

	population := createInitialPopulation(p, config.GenerationSize)

	for gen := 0; gen < config.GenerationCount; gen++ {
		fmt.Printf("Generation %d\n", gen)
	}
	sorted := sortIndividualsByFitness(population)
	for _, puzzle := range sorted {
		fmt.Println(puzzle.Fitness, puzzle.Genome)
	}
}

func Score(candidate *picross.Picross, solution *picross.Picross) int {
	score := 0

	if len(candidate.KeyDown) != len(solution.KeyDown) {
		// mismatched dimensions; treat as fully wrong
		score += len(solution.KeyDown)
	} else {
		for i := range solution.KeyDown {
			score += scoreKeys(candidate.KeyDown[i], solution.KeyDown[i])
		}
	}

	if len(candidate.KeyAcross) != len(solution.KeyAcross) {
		score += len(solution.KeyAcross)
	} else {
		for j := range solution.KeyAcross {
			score += scoreKeys(candidate.KeyAcross[j], solution.KeyAcross[j])
		}
	}

	return score
}

// scoreDim returns the number of key values that don't match the
// actual groups of true values in the row/column.
func scoreKeys(candidate, solution []int) int {
	// normalize [0] to mean "no groups"
	normalize := func(k []int) []int {
		if len(k) == 1 && k[0] == 0 {
			return nil
		}
		return k
	}
	candidate = normalize(candidate)
	solution = normalize(solution)

	maxLen := len(candidate)
	if len(solution) > maxLen {
		maxLen = len(solution)
	}

	mismatches := 0
	for i := 0; i < maxLen; i++ {
		c, s := -1, -1
		if i < len(candidate) {
			c = candidate[i]
		}
		if i < len(solution) {
			s = solution[i]
		}
		if c != s {
			mismatches++
		}
	}

	return mismatches
}

func sortIndividualsByFitness(individuals []Individual) []Individual {
	sort.Slice(individuals, func(i, j int) bool {
		return individuals[i].Fitness < individuals[j].Fitness
	})
	return individuals
}

func createInitialPopulation(p *picross.Picross, size int) []Individual {
	parents := make([]Individual, size)

	wg := sync.WaitGroup{}
	ch := make(chan Individual, size)
	for i := 0; i < size; i++ {
		wg.Go(func() {
			parent := picross.RandomPicross(p.Width, p.Height)
			score := Score(parent, p)
			ch <- Individual{Genome: parent, Fitness: score}
		})
	}
	wg.Wait()
	close(ch)
	for i := 0; i < size; i++ {
		parents[i] = <-ch
	}
	return parents
}
