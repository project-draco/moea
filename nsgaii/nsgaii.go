package nsgaii

import (
	"math"
	"sort"

	"project-draco.io/moea"
)

type NsgaIISelection struct {
	rank                  []int
	crowdingDistance      []float64
	constraintsViolations []float64
	indexes               [][]int
	pool                  []int
	elite                 []int
}

// crowddist.c: assign_crowding_distance, assign_crowding_distance_list, assign_crowding_distance_indices
// dominance.c: check_dominance
// fillnds.c: fill_nondominated_sort, crowding_fill
// rank.c: assign_ranking_and_crowding_distance
// tourselect.c: selection, tournament
// sort.c

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
	n.rank = make([]int, config.Population.Len())
	n.crowdingDistance = make([]float64, config.Population.Len()*2)
	n.constraintsViolations = make([]float64, config.Population.Len()*2)
	arr := make([]int, config.NumberOfObjectives*config.Population.Len()*2)
	n.indexes = make([][]int, config.NumberOfObjectives)
	for i := 0; i < config.NumberOfObjectives; i++ {
		n.indexes[i] = arr[i*config.Population.Len()*2 : (i+1)*config.Population.Len()*2]
	}
	n.pool = make([]int, config.Population.Len()*2)
	n.elite = make([]int, config.Population.Len()*2)
}

func (n *NsgaIISelection) assignCrowdingDistance(objectives [][]float64, dist []int) {
	if len(objectives) == 0 || len(dist) == 0 {
		return
	}
	if len(dist) <= 2 {
		n.crowdingDistance[dist[0]] = math.MaxFloat64
		if len(dist) == 2 {
			n.crowdingDistance[dist[1]] = math.MaxFloat64
		}
	}
	for i := 0; i < len(objectives[0]); i++ {
		for j := 0; j < len(dist); j++ {
			n.indexes[i][j] = dist[j]
		}
		sort.Stable(byObjectives{n.indexes[i][0:len(dist)], objectives, i})
	}
	for i := 0; i < len(dist); i++ {
		n.crowdingDistance[dist[i]] = 0.0
	}
	for i := 0; i < len(objectives[0]); i++ {
		n.crowdingDistance[n.indexes[i][0]] = math.MaxFloat64
	}
	for i := 0; i < len(objectives[0]); i++ {
		for j := 1; j < len(dist)-1; j++ {
			if n.crowdingDistance[n.indexes[i][j]] != math.MaxFloat64 &&
				objectives[n.indexes[i][len(dist)-1]][i] != objectives[n.indexes[i][0]][i] {
				n.crowdingDistance[n.indexes[i][j]] +=
					(objectives[n.indexes[i][j+1]][i] - objectives[n.indexes[i][j-1]][i]) /
						(objectives[n.indexes[i][len(dist)-1]][i] - objectives[n.indexes[i][0]][i])
			}
		}
	}
	for i := 0; i < len(dist); i++ {
		if n.crowdingDistance[dist[i]] != math.MaxFloat64 {
			n.crowdingDistance[dist[i]] /= float64(len(objectives[0]))
		}
	}
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

func (n *NsgaIISelection) crowdingFill(objectives [][]float64, mixedPopulation, newPopulation moea.Population, elite []int, start int) {
	n.assignCrowdingDistance(objectives, elite)
	for i, index := range elite {
		n.indexes[0][i] = index
	}
	sort.Stable(byDistance{n.indexes[0][0:len(elite)], n.crowdingDistance})
	for i, j := start, len(elite)-1; i < newPopulation.Len(); i, j = i+1, j-1 {
		individual := mixedPopulation.Individual(n.indexes[0][j])
		newPopulation.Individual(i).Copy(individual, 0, individual.Len())
	}
}

func (n *NsgaIISelection) fillNondominatedSort(objectives [][]float64, mixedPopulation, newPopulation moea.Population) {
	pool := n.pool[:0]
	for i := 0; i < mixedPopulation.Len(); i++ {
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
				flag = n.checkDominance(objectives, pool[j], elite[k])
				if flag == 1 {
					pool = append(pool, elite[k])
					elite = append(elite[0:k], elite[k+1:]...)
					k--
				} else if flag == -1 {
					break
				}
			}
			if flag == 0 || flag == 1 {
				elite = append(elite, pool[j])
				pool = append(pool[0:j], pool[j+1:]...)
				j--
			}
		}
		if i+len(elite) <= newPopulation.Len() {
			for _, index := range elite {
				individual := mixedPopulation.Individual(index)
				newPopulation.Individual(i).Copy(individual, 0, individual.Len())
				n.rank[i] = rank
				i++
			}
			// n.assignCrowdingDistance(objectives, elite)
			rank++
		} else {
			n.crowdingFill(objectives, mixedPopulation, newPopulation, elite, i)
			for ; i < newPopulation.Len(); i++ {
				n.rank[i] = rank
			}
		}
	}
}
