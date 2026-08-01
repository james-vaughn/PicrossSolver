package solver

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"

	"github.com/james-vaughn/PicrossSolver/picross"
)

type SolverConfig struct {
	GenerationSize  int
	GenerationCount int
	MutationRate    float64
	ElitismCount    int
	TournamentSize  int
}

type Individual struct {
	Genome  *picross.Picross
	Fitness int
}

func (i *Individual) String() string {
	return fmt.Sprintf("Genome %s\nFitness %d\n", i.Genome.String(), i.Fitness)
}

func SolvePicross(p *picross.Picross, config SolverConfig) (picross.Picross, bool) {
	// Implement the genetic algorithm to solve the Picross puzzle
	// This is a placeholder for the actual implementation

	population := createInitialPopulation(p, config.GenerationSize)

	for gen := 0; gen < config.GenerationCount; gen++ {
		fmt.Printf("Generation %d\n", gen)
		population = createGeneration(population, p, config)

		sorted := sortIndividualsByFitness(population)
		fmt.Printf("Best %s", &sorted[0])

		if sorted[0].Fitness == 0 {
			return *sorted[0].Genome, true
		}
	}
	return *sortIndividualsByFitness(population)[0].Genome, false
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
	parents := make([]Individual, 0, size)

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

	for indiv := range ch {
		parents = append(parents, indiv)
	}
	return parents
}

func createGeneration(prevGen []Individual, solution *picross.Picross, config SolverConfig) []Individual {
	genSize := len(prevGen)
	sorted := sortIndividualsByFitness(prevGen)

	nextGen := make([]Individual, 0, genSize)
	elitismCount := config.ElitismCount
	nextGen = append(nextGen, sorted[:elitismCount]...) // carry over best unchanged

	wg := sync.WaitGroup{}
	ch := make(chan Individual, genSize-elitismCount)

	for range genSize - elitismCount {
		wg.Go(func() {
			parent1 := tournamentSelect(prevGen, config.TournamentSize)
			parent2 := tournamentSelect(prevGen, config.TournamentSize)
			ch <- crossOver(parent1, parent2, solution, config.MutationRate)
		})
	}
	wg.Wait()
	close(ch)

	for indiv := range ch {
		nextGen = append(nextGen, indiv)
	}
	return nextGen
}

func tournamentSelect(generation []Individual, tournamentSize int) Individual {
	var chosen Individual

	for i := 0; i < tournamentSize; i++ {
		competitor := generation[rand.Intn(len(generation))]
		if i == 0 || competitor.Fitness < chosen.Fitness {
			chosen = competitor
		}
	}
	return chosen
}

func crossOver(p1, p2 Individual, solution *picross.Picross, mutationRate float64) Individual {
	crossOverGrid := make(picross.PicrossGrid, len(p1.Genome.Grid))
	for i, row := range p1.Genome.Grid {
		crossOverGrid[i] = append([]bool(nil), row...)
	}

	for row := 0; row < len(crossOverGrid); row++ {
		for col := 0; col < len(crossOverGrid[0]); col++ {
			if rand.Float64() < 0.5 {
				crossOverGrid[row][col] = p2.Genome.Grid[row][col]
			}
		}
	}
	mutate(&crossOverGrid, mutationRate)

	child := picross.FromGrid(crossOverGrid)
	return Individual{
		Genome:  child,
		Fitness: Score(child, solution),
	}
}

func mutate(grid *picross.PicrossGrid, mutationRate float64) {
	for row := 0; row < len(*grid); row++ {
		for col := 0; col < len((*grid)[0]); col++ {
			if rand.Float64() < mutationRate {
				(*grid)[row][col] = !(*grid)[row][col]
			}
		}
	}

}
