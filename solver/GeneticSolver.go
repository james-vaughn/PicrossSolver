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
	mutationRate := config.MutationRate
	for gen := 0; gen < config.GenerationCount; gen++ {
		fmt.Printf("Generation %d\n", gen)
		population = createGeneration(population, p, config, mutationRate)

		sorted := sortIndividualsByFitness(population)
		fmt.Printf("Best Fitness %d\nMutation Rate %f\n", sorted[0].Fitness, mutationRate)

		mutationRate = adaptiveMutationRate(config.MutationRate, sorted[0].Fitness, len(p.Grid)*len(p.Grid[0])/2)
		if sorted[0].Fitness == 0 {
			return *sorted[0].Genome, true
		}
	}
	return *sortIndividualsByFitness(population)[0].Genome, false
}

func Score(candidate *picross.Picross, solution *picross.Picross) int {
	score := 0

	if candidate.Height != len(solution.KeyDown) {
		score += len(solution.KeyDown)
	} else {
		for i := range solution.KeyDown {
			score += lineDistance(candidate.Grid[i], solution.KeyDown[i])
		}
	}

	if candidate.Width != len(solution.KeyAcross) {
		score += len(solution.KeyAcross)
	} else {
		for j := range solution.KeyAcross {
			column := make([]bool, candidate.Height)
			for i := 0; i < candidate.Height; i++ {
				column[i] = candidate.Grid[i][j]
			}
			score += lineDistance(column, solution.KeyAcross[j])
		}
	}

	return score
}

// lineDistance returns the minimum number of cell flips needed to turn
// `row` into some row that exactly satisfies `key`.
func lineDistance(row []bool, key []int) int {
	groups := key
	if len(key) == 1 && key[0] == 0 {
		groups = nil
	}
	n := len(row)
	m := len(groups)

	// prefix[i] = cost of making cells [0,i) filled, relative to `row`
	prefix := make([]int, n+1)
	for i := 0; i < n; i++ {
		c := 0
		if !row[i] {
			c = 1
		}
		prefix[i+1] = prefix[i] + c
	}
	filledCost := func(start, length int) int {
		return prefix[start+length] - prefix[start]
	}
	emptyCost := func(i int) int {
		if row[i] {
			return 1
		}
		return 0
	}

	const inf = 1 << 30
	// dp[pos][j][s]: min cost using first `pos` cells, `j` groups placed,
	// s=0 free to start next group, s=1 just finished a group (needs a gap)
	dp := make([][][2]int, n+1)
	for i := range dp {
		dp[i] = make([][2]int, m+1)
		for j := range dp[i] {
			dp[i][j][0], dp[i][j][1] = inf, inf
		}
	}
	dp[0][0][0] = 0

	for i := 0; i < n; i++ {
		for j := 0; j <= m; j++ {
			for s := 0; s < 2; s++ {
				cur := dp[i][j][s]
				if cur == inf {
					continue
				}
				// place an empty cell here
				if cost := cur + emptyCost(i); cost < dp[i+1][j][0] {
					dp[i+1][j][0] = cost
				}
				// start the next group here (only if free to, i.e. s==0)
				if s == 0 && j < m {
					k := groups[j]
					if i+k <= n {
						if cost := cur + filledCost(i, k); cost < dp[i+k][j+1][1] {
							dp[i+k][j+1][1] = cost
						}
					}
				}
			}
		}
	}

	best := dp[n][m][0]
	if dp[n][m][1] < best {
		best = dp[n][m][1]
	}
	return best
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

func createGeneration(prevGen []Individual, solution *picross.Picross, config SolverConfig, mutationRate float64) []Individual {
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
			ch <- crossOver(parent1, parent2, solution, mutationRate)
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

// reduce mutation rate as the best fitness approaches zero, to avoid overshooting the solution
func adaptiveMutationRate(baseRate float64, bestFitness int, worstCaseFitness int) float64 {
	progress := 1.0 - float64(bestFitness)/float64(worstCaseFitness)
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	return baseRate * (1.0 - 0.9*progress)
}
