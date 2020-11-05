package nsgaii

import (
	"math"
	"sort"

	"github.com/project-draco/moea"
)

type NsgaIISelection struct {
	Rank                  []int
	crowdingDistance      []float64
	mixedCrowdingDistance []float64
	constraintsViolations []float64
	previousPopulation    moea.Population
	previousObjectives    [][]float64
	mixedPopulation       mixedPopulation
	mixedObjectives       [][]float64
	indexes               [][]int
	pool                  []int
	elite                 []int
	sequence              []int
}

// crowddist.c: assign_crowding_distance, assign_crowding_distance_list, assign_crowding_distance_indices
// dominance.c: check_dominance
// fillnds.c: fill_nondominated_sort, crowding_fill
// rank.c: assign_ranking_and_crowding_distance
// tourselect.c: selection, tournament
// sort.c

type mixedPopulation []moea.Individual

func (m mixedPopulation) Len() int {
	return len(m)
}

func (m mixedPopulation) Individual(i int) moea.Individual {
	return m[i]
}

func (m mixedPopulation) Clone() moea.Population {
	return nil
}

type byObjectives struct {
	indexes         []int
	objectives      [][]float64
	objectivesIndex int
}

func (bo byObjectives) Len() int {
	return len(bo.indexes)
}

func (bo byObjectives) Swap(i, j int) {
	bo.indexes[i], bo.indexes[j] = bo.indexes[j], bo.indexes[i]
}

func (bo byObjectives) Less(i, j int) bool {
	return bo.objectives[bo.indexes[i]][bo.objectivesIndex] <
		bo.objectives[bo.indexes[j]][bo.objectivesIndex]
}

type byDistance struct {
	indexes   []int
	distances []float64
}

func (bd byDistance) Len() int {
	return len(bd.indexes)
}

func (bd byDistance) Swap(i, j int) {
	bd.indexes[i], bd.indexes[j] = bd.indexes[j], bd.indexes[i]
}

func (bd byDistance) Less(i, j int) bool {
	return bd.distances[bd.indexes[i]] < bd.distances[bd.indexes[j]]
}

func (n *NsgaIISelection) Initialize(config *moea.Config) {
	n.Rank = make([]int, config.Population.Len())
	n.crowdingDistance = make([]float64, config.Population.Len())
	n.mixedCrowdingDistance = make([]float64, config.Population.Len()*2)
	n.constraintsViolations = make([]float64, config.Population.Len()*2)
	n.mixedPopulation = make(mixedPopulation, config.Population.Len()*2)
	clone1 := config.Population.Clone()
	clone2 := config.Population.Clone()
	for i := 0; i < config.Population.Len()*2; i += 2 {
		n.mixedPopulation[i] = clone1.Individual(i / 2)
		n.mixedPopulation[i+1] = clone2.Individual(i / 2)
	}
	n.mixedObjectives = make([][]float64, config.Population.Len()*2)
	arr := make([]int, config.NumberOfObjectives*config.Population.Len()*2)
	n.indexes = make([][]int, config.NumberOfObjectives)
	for i := 0; i < config.NumberOfObjectives; i++ {
		n.indexes[i] = arr[i*config.Population.Len()*2 : (i+1)*config.Population.Len()*2]
	}
	n.pool = make([]int, config.Population.Len()*2)
	n.elite = make([]int, config.Population.Len()*2)
	n.sequence = make([]int, config.Population.Len())
	for i := 0; i < config.Population.Len(); i++ {
		n.sequence[i] = i
	}
}

func (n *NsgaIISelection) OnGeneration(config *moea.Config, population moea.Population, objectives [][]float64) {
	if n.previousPopulation == nil {
		n.assignRankAndCrowdingDistance(objectives)
	} else {
		n.merge(population, objectives)
		n.fillNondominatedSort(population, objectives)
	}
	n.previousPopulation = population
	n.previousObjectives = objectives
}

func (n *NsgaIISelection) Finalize(config *moea.Config, population moea.Population, objectives [][]float64, result *moea.Result) {
	n.merge(population, objectives)
	n.fillNondominatedSort(population, objectives)
	for i := 0; i < population.Len(); i++ {
		result.Individuals[i].Objective = objectives[i]
		for j := 0; j < config.NumberOfValues; j++ {
			result.Individuals[i].Values[j] = population.Individual(i).Value(j)
		}
		result.Individuals[i].Parent1 = -1
		result.Individuals[i].Parent2 = -1
		result.Individuals[i].CrossSite = -1
		if result.BestObjective[0] > result.Individuals[i].Objective[0] {
			result.BestObjective[0] = result.Individuals[i].Objective[0]
			result.BestIndividual = population.Individual(i)
			result.BestIndividualIndex = i
		}
	}
}

func (n *NsgaIISelection) Selection(config *moea.Config, objectives [][]float64) int {
	r0 := int(config.RandomNumberGenerator.Float64() * float64(config.Population.Len()-1))
	r1 := int(config.RandomNumberGenerator.Float64() * float64(config.Population.Len()-1))
	flag := n.checkDominance(objectives, r0, r1)
	if flag == 1 {
		return r0
	} else if flag == -1 {
		return r1
	} else if n.crowdingDistance[r0] > n.crowdingDistance[r1] {
		return r0
	} else if n.crowdingDistance[r1] > n.crowdingDistance[r0] {
		return r1
	} else if config.RandomNumberGenerator.FairFlip() {
		return r0
	}
	return r1
}

func (n *NsgaIISelection) checkDominance(objectives [][]float64, a, b int) int {
	if n.constraintsViolations[a] < 0 && n.constraintsViolations[b] < 0 {
		if n.constraintsViolations[a] > n.constraintsViolations[b] {
			return 1
		} else if n.constraintsViolations[a] < n.constraintsViolations[b] {
			return -1
		} else {
			return 0
		}
	} else if n.constraintsViolations[a] < 0 && n.constraintsViolations[b] == 0 {
		return -1
	} else if n.constraintsViolations[a] == 0 && n.constraintsViolations[b] < 0 {
		return 1
	} else {
		flag1 := false
		flag2 := false
		for i := 0; i < len(objectives[a]); i++ {
			if objectives[a][i] < objectives[b][i] {
				flag1 = true
			} else if objectives[a][i] > objectives[b][i] {
				flag2 = true
			}
		}
		if flag1 && !flag2 {
			return 1
		} else if !flag1 && flag2 {
			return -1
		} else {
			return 0
		}
	}
}

func (n *NsgaIISelection) assignCrowdingDistance(objectives [][]float64, dist []int, crowdingDistance []float64) {
	if len(objectives) == 0 || len(dist) == 0 {
		return
	}
	if len(dist) <= 2 {
		crowdingDistance[dist[0]] = math.MaxFloat64
		if len(dist) == 2 {
			crowdingDistance[dist[1]] = math.MaxFloat64
		}
	}
	for i := 0; i < len(objectives[0]); i++ {
		for j := 0; j < len(dist); j++ {
			n.indexes[i][j] = dist[j]
		}
		sort.Stable(byObjectives{n.indexes[i][0:len(dist)], objectives, i})
	}
	for i := 0; i < len(dist); i++ {
		crowdingDistance[dist[i]] = 0.0
	}
	for i := 0; i < len(objectives[0]); i++ {
		crowdingDistance[n.indexes[i][0]] = math.MaxFloat64
	}
	for i := 0; i < len(objectives[0]); i++ {
		for j := 1; j < len(dist)-1; j++ {
			if crowdingDistance[n.indexes[i][j]] != math.MaxFloat64 &&
				objectives[n.indexes[i][len(dist)-1]][i] != objectives[n.indexes[i][0]][i] {
				crowdingDistance[n.indexes[i][j]] +=
					(objectives[n.indexes[i][j+1]][i] - objectives[n.indexes[i][j-1]][i]) /
						(objectives[n.indexes[i][len(dist)-1]][i] - objectives[n.indexes[i][0]][i])
			}
		}
	}
	for i := 0; i < len(dist); i++ {
		if crowdingDistance[dist[i]] != math.MaxFloat64 {
			crowdingDistance[dist[i]] /= float64(len(objectives[0]))
		}
	}
}

func (n *NsgaIISelection) crowdingFill(newPopulation moea.Population, newObjectives [][]float64, elite []int, start int) {
	n.assignCrowdingDistance(n.mixedObjectives, elite, n.mixedCrowdingDistance)
	for i, index := range elite {
		n.indexes[0][i] = index
	}
	sort.Stable(byDistance{n.indexes[0][0:len(elite)], n.mixedCrowdingDistance})
	for i, j := start, len(elite)-1; i < newPopulation.Len(); i, j = i+1, j-1 {
		individual := n.mixedPopulation.Individual(n.indexes[0][j])
		newPopulation.Individual(i).Copy(individual, 0, individual.Len())
		newObjectives[i] = n.mixedObjectives[n.indexes[0][j]]
		n.crowdingDistance[i] = n.mixedCrowdingDistance[n.indexes[0][j]]
	}
}

func (n *NsgaIISelection) fillNondominatedSort(newPopulation moea.Population, newObjectives [][]float64) {
	pool := n.pool[:0]
	for i := 0; i < n.mixedPopulation.Len(); i++ {
		pool = append(pool, i)
	}
	rank := 1
	for i := 0; i < newPopulation.Len(); {
		elite := n.elite[0:1]
		elite[0] = pool[0]
		pool = pool[1:]
		for j := 0; j < len(pool); j++ {
			var flag int
			for k := 0; k < len(elite); k++ {
				flag = n.checkDominance(n.mixedObjectives, pool[j], elite[k])
				if flag == 1 {
					pool = append(pool, elite[k])
					elite = append(elite[:k], elite[k+1:]...)
					k--
				} else if flag == -1 {
					break
				}
			}
			if flag == 0 || flag == 1 {
				elite = append(elite, pool[j])
				pool = append(pool[:j], pool[j+1:]...)
				j--
			}
		}
		if i+len(elite) <= newPopulation.Len() {
			j := i
			for _, index := range elite {
				individual := n.mixedPopulation.Individual(index)
				newPopulation.Individual(i).Copy(individual, 0, individual.Len())
				newObjectives[i] = n.mixedObjectives[index]
				n.Rank[i] = rank
				i++
			}
			n.assignCrowdingDistance(newObjectives, n.sequence[j:j+len(elite)], n.crowdingDistance)
			rank++
		} else {
			n.crowdingFill(newPopulation, newObjectives, elite, i)
			for ; i < newPopulation.Len(); i++ {
				n.Rank[i] = rank
			}
		}
	}
}

func (n *NsgaIISelection) assignRankAndCrowdingDistance(objectives [][]float64) {
	orig := n.pool[:0]
	for i := 0; i < len(objectives); i++ {
		orig = append(orig, i)
	}
	rank := 1
	for len(orig) > 0 {
		if len(orig) == 1 {
			n.Rank[orig[0]] = rank
			n.crowdingDistance[orig[0]] = math.MaxFloat64
			break
		}
		cur := n.elite[:1]
		cur[0] = orig[0]
		orig = orig[1:]
		for i := 0; i < len(orig); i++ {
			var flag int
			for j := 0; j < len(cur); j++ {
				flag = n.checkDominance(objectives, orig[i], cur[j])
				if flag == 1 {
					orig = append(orig, cur[j])
					cur = append(cur[:j], cur[j+1:]...)
					j--
				} else if flag == -1 {
					break
				}
			}
			if flag == 0 || flag == 1 {
				cur = append(cur, orig[i])
				orig = append(orig[:i], orig[i+1:]...)
				i--
			}
		}
		for i := 0; i < len(cur); i++ {
			n.Rank[cur[i]] = rank
		}
		n.assignCrowdingDistance(objectives, cur, n.crowdingDistance)
		rank++
	}
}

func (n *NsgaIISelection) merge(population moea.Population, objectives [][]float64) {
	for i := 0; i < n.previousPopulation.Len()*2; i++ {
		var individual moea.Individual
		if i < n.previousPopulation.Len() {
			individual = n.previousPopulation.Individual(i)
		} else {
			individual = population.Individual(i - n.previousPopulation.Len())
		}
		n.mixedPopulation[i].Copy(individual, 0, individual.Len())
	}
	for i, o := range n.previousObjectives {
		n.mixedObjectives[i] = o
	}
	for i, o := range objectives {
		n.mixedObjectives[i+len(objectives)] = o
	}
}
