package nsgaiii

import (
	"math"
	"sort"
	"fmt"
	"../"
)

type NsgaIIISelection struct {
	ReferencePointsDivision int
	Rank                  []int
	referencePointArray   []ReferencePoint
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

type ReferencePoint struct {
	position []float64
}

type NormalizedIndividual struct {
	index int
	objectives []float64
}


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

func generateReferencePoints(numberOfDivisions int, nroObjectives int) []ReferencePoint {
	var referencePointArray []ReferencePoint
	var referencePoint ReferencePoint
	referencePoint.position = make([]float64, nroObjectives)
	generateReferencePointsRecursive(&referencePointArray, referencePoint, nroObjectives, numberOfDivisions, numberOfDivisions, 0)
	return referencePointArray
}

func generateReferencePointsRecursive(referencePointArray *[]ReferencePoint, currentPoint ReferencePoint, numberOfObjectives int, left int, total int, element int) {
	if(element == (numberOfObjectives - 1)) {
		currentPoint.position[element] = float64(left)/float64(total)
		*referencePointArray = append(*referencePointArray, currentPoint)
	} else {
		for i := 0; i <= left; i++ {
			currentPoint.position[element] = float64(i)/float64(total)
			generateReferencePointsRecursive(referencePointArray, currentPoint, numberOfObjectives, left - i, total, element + 1)
		}
	}
}

func (n *NsgaIIISelection) Initialize(config *moea.Config) {
	n.referencePointArray = generateReferencePoints(n.ReferencePointsDivision, config.NumberOfObjectives)
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

func (n *NsgaIIISelection) OnGeneration(config *moea.Config, population moea.Population, objectives [][]float64) {
	if n.previousPopulation == nil {
		//fmt.Printf("ehhhh\n")
		n.assignRankAndCrowdingDistance(objectives)
		/*for i := 0; i < len(objectives); i++ {
			fmt.Printf("[")
			for j := 0; j < config.NumberOfObjectives; j++ {
				fmt.Printf(" %.4f,", objectives[i][j])
			}
			fmt.Printf("]\n")
		}*/
	} else {
		//fmt.Printf("phhhh\n")
		n.merge(population, objectives)
		n.fillNondominatedSort(population, objectives)
		/*for i := 0; i < len(objectives); i++ {
			fmt.Printf("[")
			for j := 0; j < config.NumberOfObjectives; j++ {
				fmt.Printf(" %.4f,", objectives[i][j])
			}
			fmt.Printf("]\n")
		}*/
	}
	n.previousPopulation = population
	n.previousObjectives = objectives
}

func (n *NsgaIIISelection) Finalize(config *moea.Config, population moea.Population, objectives [][]float64, result *moea.Result) {
	//fmt.Printf("aqui3\n")
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
		}
	}
}

func (n *NsgaIIISelection) Selection(config *moea.Config, objectives [][]float64) int {
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

func (n *NsgaIIISelection) checkDominance(objectives [][]float64, a, b int) int {
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

func (n *NsgaIIISelection) assignCrowdingDistance(objectives [][]float64, dist []int, crowdingDistance []float64) {
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

func (n *NsgaIIISelection) crowdingFill(newPopulation moea.Population, newObjectives [][]float64, elite []int, start int) {
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

func  (n *NsgaIIISelection) normalizeObjectives(individualsIndexes []int, intercepts []float64, idealPoint[]float64, nroObjectives int) []NormalizedIndividual{
	var normalizedIndividuals = make([]NormalizedIndividual, 0)
	for i := 0; i < len(individualsIndexes); i++ {
		var normalizeObjectives = make([]float64, nroObjectives)
		for j := 0; j < nroObjectives; j++ {
			normalizeObjectives[j] = normalizeObjective(n.mixedObjectives[individualsIndexes[i]], j, intercepts, idealPoint)
		}
		var normalizedIndividual NormalizedIndividual
		normalizedIndividual.objectives = normalizeObjectives
		normalizedIndividual.index = individualsIndexes[i]
		normalizedIndividuals = append(normalizedIndividuals, normalizedIndividual)
	}
	return normalizedIndividuals
}

func normalizeObjective(individual []float64, objectiveIndex int, intercepts []float64,  idealPoint[]float64) float64 {
	var epsilon = 1e-20
	if math.Abs(intercepts[objectiveIndex] - idealPoint[objectiveIndex]) > epsilon {
		return individual[objectiveIndex] / (intercepts[objectiveIndex] - idealPoint[objectiveIndex])
	} else {
		return individual[objectiveIndex] / epsilon
	}
}

func minimunArray(a []float64, b []float64)[]float64 {
	var aux = make([]float64, len(a))
	for i := 0; i < len(aux); i++ {
		if(a[i] < b[i]) {
			aux[i] = a[i]
		} else {
			aux[i] = b[i]
		}
	}
	return aux
}

func (n *NsgaIIISelection) findIdealPoint(elite []int, nroObjectives int) []float64{
	var currentIdeal = make([]float64, nroObjectives)
	for i := 0; i < len(currentIdeal); i++ {
		currentIdeal[i] = math.Inf(1)
	}
	for _, index := range elite {
		var aux = make([]float64, nroObjectives)
		for i := 0; i < len(aux); i++ {
			aux[i] = n.mixedObjectives[index][i] * -1
		}
		currentIdeal = minimunArray(currentIdeal, aux)
	}
	return currentIdeal
}

func (n *NsgaIIISelection) findExtremeIndividualForObjective(elite []int, objective int) int {
	var maxValue = math.Inf(-1)
	var maxValueIndex = 0
	for i := 0; i < len(elite); i++ {
		if(n.mixedObjectives[elite[i]][objective] * -1 > maxValue) {
			maxValue = n.mixedObjectives[elite[i]][objective] * -1
			maxValueIndex = elite[i]
		}
	}
	return maxValueIndex
}

func (n *NsgaIIISelection) findExtremePoints(elite []int, nroObjectives int) []int {
	var extremePoints = make([]int, nroObjectives)
	for i := 0; i < len(extremePoints); i++ {
		var index = n.findExtremeIndividualForObjective(elite, i)
		extremePoints[i] = index
	}
	return extremePoints
}

func (n *NsgaIIISelection) guassianElimination(a [][]float64, b []float64) []float64 {
	var N = len(a)
	var x = make([]float64, N)
	for i := 0; i < N; i++ {
		a[i] = append(a[i], b[i])
	}
	for base := 0; base < N-1; base++ {
		for target := base+1; target < N; target++ {
			var ratio = a[target][base]/a[base][base]
			for term := 0; term < len(a[base]); term++ {
				a[target][term] = a[target][term] - a[base][term]*ratio;
			}
		}
	}
	for i := 0; i < len(x); i++ {
		x[i] = 0
	}
	for i := N-1; i >=0 ; i--{
		for known := i+1; known<N; known++ {
			a[i][N] = a[i][N] - a[i][known]*x[known]
		}
		x[i] = a[i][N] / a[i][i]
	}
	return x
}

func (n *NsgaIIISelection) constructHyperplane(elite []int, extremes []int, nroObjectives int) []float64{
	var intercepts = make([]float64, nroObjectives)
	if(n.hasDuplicateIndividuals(elite)) {
		for i := 0; i < len(intercepts); i++ {
			intercepts[i] = n.mixedObjectives[extremes[i]][i]
		}
	} else {
		var b = make([]float64, nroObjectives)
		for i := 0; i < len(b); i++ {
			b[i] = 1
		}
		var a = make([][]float64, len(extremes))
		for i := 0; i < len(a); i++ {
			var aux = make([]float64, nroObjectives)
			for j := 0; j < nroObjectives; j++ {
				aux[j] = n.mixedObjectives[extremes[i]][j]
			}
			a[i] = aux
		}
		var x = n.guassianElimination(a, b)
		for i := 0; i < len(intercepts); i++ {
			intercepts[i] = 1/x[i]
		}
	}
	return intercepts
}

func (n *NsgaIIISelection) hasDuplicateIndividuals(elite []int) bool{
	for i := 0; i < len(elite); i++ {
		for j := 0; j < len(elite); j++ {
			if(j != i) {
				if(n.hasSameValuesForObjectives(n.mixedObjectives[elite[i]], n.mixedObjectives[elite[j]])) {
					return true
				}
			}
		}
	}
	return false
}

func (n *NsgaIIISelection) hasSameValuesForObjectives(a []float64, b []float64) bool{
	for i := 0; i < len(a); i++ {
		if(a[i] == b[i]) {
			return false
		}
	}
	return true
}

func (n *NsgaIIISelection) fillNondominatedSort(newPopulation moea.Population, newObjectives [][]float64) {
	pool := n.pool[:0]
	for i := 0; i < n.mixedPopulation.Len(); i++ {
		pool = append(pool, i)
		//fmt.Printf("%d, ", pool[i])
	}
	//fmt.Printf("\n")
	remaining := newPopulation.Len()
	rank := 1
	for i := 0; i < newPopulation.Len(); {
		elite := n.elite[0:1]
		elite[0] = pool[0]
		pool = pool[1:]
		for j := 0; j < len(pool); j++ {
			var flag int
			//fmt.Printf("----> %d\n", elite)
			for k := 0; k < len(elite); k++ {
				flag = n.checkDominance(n.mixedObjectives, pool[j], elite[k])
				//fmt.Printf("----> %d\n", flag)
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
				//fmt.Printf("----> %d\n", elite)
				pool = append(pool[:j], pool[j+1:]...)
				//fmt.Printf("----> %d\n", pool)
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
			remaining -= len(elite)
			n.assignCrowdingDistance(newObjectives, n.sequence[j:j+len(elite)], n.crowdingDistance)
			rank++
		} else {
			//n.crowdingFill(newPopulation, newObjectives, elite, i)
			var nroObjectives = len(n.mixedObjectives[0])
			var idealPoint = n.findIdealPoint(elite, nroObjectives)
			var extremes = n.findExtremePoints(elite, nroObjectives)
			var intercepts = n.constructHyperplane(elite, extremes, nroObjectives)
			var normalizedIndividuals = n.normalizeObjectives(elite, intercepts, idealPoint, nroObjectives)
			for ; i < newPopulation.Len(); i++ {
				n.Rank[i] = rank
			}
		}
	}
}

func (n *NsgaIIISelection) assignRankAndCrowdingDistance(objectives [][]float64) {
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

func (n *NsgaIIISelection) merge(population moea.Population, objectives [][]float64) {
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
