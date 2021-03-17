package nsgaiii

import (
	"math"
	"math/rand"
	"sort"

	"github.com/JoaoGabriel0511/moea"
)

type NsgaIIISelection struct {
	ReferencePointsDivision int
	referencePointArray     []ReferencePoint
	mixedObjectives         [][]float64
	mixedPopulation         mixedPopulation
}

type mixedPopulation []moea.Individual
type ReferencePoint struct {
	position         []float64
	associationCount int
	associations     []NormalizedIndividual
}

type NormalizedIndividual struct {
	index          int
	objectives     []float64
	referencePoint ReferencePoint
	distance       float64
}

type Association struct {
	point *ReferencePoint
	dist  float64
}

func generateReferencePoints(numberOfDivisions int, nroObjectives int) []ReferencePoint {
	var referencePointArray []ReferencePoint
	var referencePoint ReferencePoint
	referencePoint.position = make([]float64, nroObjectives)
	generateReferencePointsRecursive(&referencePointArray, referencePoint, nroObjectives, numberOfDivisions, numberOfDivisions, 0)
	return referencePointArray
}

func generateReferencePointsRecursive(referencePointArray *[]ReferencePoint, currentPoint ReferencePoint, numberOfObjectives int, left int, total int, element int) {
	if element == (numberOfObjectives - 1) {
		currentPoint.position[element] = float64(left) / float64(total)
		var referencePoint = currentPoint
		referencePoint.position = make([]float64, len(currentPoint.position))
		for i := 0; i < len(referencePoint.position); i++ {
			referencePoint.position[i] = currentPoint.position[i]
		}
		referencePoint.associations = make([]NormalizedIndividual, 0)
		referencePoint.associationCount = 0
		*referencePointArray = append(*referencePointArray, referencePoint)
	} else {
		for i := 0; i <= left; i++ {
			currentPoint.position[element] = float64(i) / float64(total)
			generateReferencePointsRecursive(referencePointArray, currentPoint, numberOfObjectives, left-i, total, element+1)
		}
	}
}

func (n *NsgaIIISelection) SelectRemaining(remaining int, elite []int, mixedObjectives [][]float64, mixedPopulation []moea.Individual) []int {
	n.mixedObjectives = mixedObjectives
	n.mixedPopulation = mixedPopulation
	var nroObjectives = len(n.mixedObjectives[0])
	var idealPoint = n.findIdealPoint(elite, nroObjectives)
	var extremes = n.findExtremePoints(elite, nroObjectives)
	var intercepts = n.constructHyperplane(elite, extremes, nroObjectives)
	var normalizedIndividuals = n.normalizeObjectives(elite, intercepts, idealPoint, nroObjectives)
	n.associate(normalizedIndividuals, nroObjectives)
	var result = make([]int, 0)
	for len(result) < remaining {
		var minAssocRps = getMinimalAssociationCountReferences(n.referencePointArray)
		if len(minAssocRps) > 0 {
			var chosenRPIndex = rand.Intn(len(minAssocRps))
			var chosenRP = n.referencePointArray[minAssocRps[chosenRPIndex]]
			if len(chosenRP.associations) != 0 {
				var selected NormalizedIndividual
				var selectedIndex int
				if chosenRP.associationCount != 0 {
					selectedIndex = getMinimalRefPointDistIndividual(chosenRP.associations)
				} else {
					selectedIndex = rand.Intn(len(chosenRP.associations))
				}
				selected = chosenRP.associations[selectedIndex]
				result = append(result, selected.index)
				chosenRP.associationCount++
				chosenRP.associations = append(chosenRP.associations[:selectedIndex], chosenRP.associations[selectedIndex+1:]...)
			} else {
				n.referencePointArray = append(n.referencePointArray[:minAssocRps[chosenRPIndex]], n.referencePointArray[minAssocRps[chosenRPIndex]+1:]...)
			}
		}
	}
	for i := 0; i < len(n.referencePointArray); i++ {
		n.referencePointArray[i].associationCount = 0
		n.referencePointArray[i].associations = nil
	}
	return result
}

func (n *NsgaIIISelection) Initialize(numberOfObjectives int) {
	n.referencePointArray = generateReferencePoints(n.ReferencePointsDivision, numberOfObjectives)
}

func (n *NsgaIIISelection) normalizeObjectives(individualsIndexes []int, intercepts []float64, idealPoint []float64, nroObjectives int) []NormalizedIndividual {
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

func normalizeObjective(individual []float64, objectiveIndex int, intercepts []float64, idealPoint []float64) float64 {
	var epsilon = 1e-20
	if math.Abs(intercepts[objectiveIndex]-idealPoint[objectiveIndex]) > epsilon {
		return individual[objectiveIndex] / (intercepts[objectiveIndex] - idealPoint[objectiveIndex])
	} else {
		return individual[objectiveIndex] / epsilon
	}
}

func minimunArray(a []float64, b []float64) []float64 {
	var aux = make([]float64, len(a))
	for i := 0; i < len(aux); i++ {
		if a[i] < b[i] {
			aux[i] = a[i]
		} else {
			aux[i] = b[i]
		}
	}
	return aux
}

func (n *NsgaIIISelection) findIdealPoint(elite []int, nroObjectives int) []float64 {
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
		if n.mixedObjectives[elite[i]][objective]*-1 > maxValue {
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

func (n *NsgaIIISelection) constructHyperplane(elite []int, extremes []int, nroObjectives int) []float64 {
	var intercepts = make([]float64, nroObjectives)
	if n.hasDuplicateIndividuals(elite) {
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
			intercepts[i] = 1 / x[i]
		}
	}
	return intercepts
}

func (n *NsgaIIISelection) guassianElimination(a [][]float64, b []float64) []float64 {
	var N = len(a)
	var x = make([]float64, N)
	for i := 0; i < N; i++ {
		a[i] = append(a[i], b[i])
	}
	for base := 0; base < N-1; base++ {
		for target := base + 1; target < N; target++ {
			var ratio = a[target][base] / a[base][base]
			for term := 0; term < len(a[base]); term++ {
				a[target][term] = a[target][term] - a[base][term]*ratio
			}
		}
	}
	for i := 0; i < len(x); i++ {
		x[i] = 0
	}
	for i := N - 1; i >= 0; i-- {
		for known := i + 1; known < N; known++ {
			a[i][N] = a[i][N] - a[i][known]*x[known]
		}
		x[i] = a[i][N] / a[i][i]
	}
	return x
}

func (n *NsgaIIISelection) hasDuplicateIndividuals(elite []int) bool {
	for i := 0; i < len(elite); i++ {
		for j := 0; j < len(elite); j++ {
			if j != i {
				if n.hasSameValuesForObjectives(n.mixedObjectives[elite[i]], n.mixedObjectives[elite[j]]) {
					return true
				}
			}
		}
	}
	return false
}

func (n *NsgaIIISelection) hasSameValuesForObjectives(a []float64, b []float64) bool {
	for i := 0; i < len(a); i++ {
		if a[i] == b[i] {
			return false
		}
	}
	return true
}

func perpendicularDistance(normalizedObjectives []float64, referencePoint []float64) float64 {
	var numerator float64 = 0
	var denominator float64 = 0
	for i := 0; i < len(referencePoint); i++ {
		numerator += referencePoint[i] * normalizedObjectives[i]
		denominator += referencePoint[i] * referencePoint[i]
	}
	var k = numerator / denominator
	var d float64 = 0
	for i := 0; i < len(referencePoint); i++ {
		d += (k*referencePoint[i] - normalizedObjectives[i]) * (k*referencePoint[i] - normalizedObjectives[i])
	}
	return math.Sqrt(d)
}

func (n *NsgaIIISelection) associate(normalizedIndividuals []NormalizedIndividual, nroObjectives int) {
	for index, individual := range normalizedIndividuals {
		var rpDist = make([]Association, len(n.referencePointArray))
		for i := 0; i < len(n.referencePointArray); i++ {
			var association Association
			association.point = &n.referencePointArray[i]
			association.dist = perpendicularDistance(individual.objectives, n.referencePointArray[i].position)
			rpDist[i] = association
		}
		sort.SliceStable(rpDist, func(i, j int) bool {
			return rpDist[i].dist < rpDist[j].dist
		})
		var bestDist = rpDist[0].dist
		var bestRp = rpDist[0].point
		normalizedIndividuals[index].referencePoint = *bestRp
		normalizedIndividuals[index].distance = bestDist
		bestRp.associationCount++
		bestRp.associations = append(bestRp.associations, individual)
	}
}

func getMinimalRefPointDistIndividual(individuals []NormalizedIndividual) int {
	var minDist = math.Inf(1)
	var index int
	for i := 0; i < len(individuals); i++ {
		if individuals[i].distance < minDist {
			index = i
			minDist = individuals[i].distance
		}
	}
	return index
}

func getMinimalAssociationCountReferences(referencePoints []ReferencePoint) []int {
	var minCount = math.Inf(1)
	var res = make([]int, 0)
	for i := 0; i < len(referencePoints); i++ {
		if float64(referencePoints[i].associationCount) < minCount {
			minCount = float64(referencePoints[i].associationCount)
		}
	}
	for i := 0; i < len(referencePoints); i++ {
		if referencePoints[i].associationCount == int(minCount) {
			res = append(res, i)
		}
	}
	return res
}
