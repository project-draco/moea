package nsga

import (
	"math"

	"project-draco.io/moea"
)

type NsgaSelection struct {
	Variables   func(i int) []float64
	LowerBounds []float64
	UpperBounds []float64
	Dshare      float64
	front       []int
	flag        []int
	dumfitness  []float64
	mindum      float64
	deltadum    float64
	choices     []int
	fraction    []float64
	nremain     int
}

func (ns *NsgaSelection) initialize(config *moea.Config) {
	ns.front = make([]int, config.Population.Len())
	ns.flag = make([]int, config.Population.Len())
	ns.dumfitness = make([]float64, config.Population.Len())
	ns.deltadum = 0.1 * float64(config.Population.Len())
	ns.choices = make([]int, config.Population.Len())
	ns.fraction = make([]float64, config.Population.Len())
	if ns.Dshare == 0 {
		ns.Dshare = 0.1
	}
}

func (ns *NsgaSelection) onGeneration(config *moea.Config, objectives [][]float64) {
	popcount := 0
	frontindex := 1
	for i := 0; i < config.Population.Len(); i++ {
		ns.flag[i] = 0
		ns.dumfitness[i] = 0.0
	}
	for popcount < config.Population.Len() {
		for i := 0; i < config.Population.Len(); i++ {
			if ns.flag[i] == 3 {
				continue
			}
			for j := 0; j < config.Population.Len(); j++ {
				if i == j {
					continue
				} else if ns.flag[j] == 3 {
					continue
				} else if ns.flag[j] == 1 {
					continue
				} else {
					flagobj := false
					for k := 0; k < len(objectives[i]); k++ {
						if objectives[i][k] >= objectives[j][k] {
							flagobj = true
							break
						}
					}
					if !flagobj {
						ns.flag[i] = 1
						break
					}
				}
			}
			if ns.flag[i] == 0 {
				ns.flag[i] = 2
				popcount++
			}
		}
		if frontindex == 1 {
			for i := 0; i < config.Population.Len(); i++ {
				if ns.flag[i] == 2 {
					ns.dumfitness[i] = float64(config.Population.Len())
					ns.front[i] = frontindex
				}
			}
		} else {
			for i := 0; i < config.Population.Len(); i++ {
				if ns.flag[i] == 2 {
					ns.front[i] = frontindex
					if ns.mindum > ns.deltadum {
						ns.dumfitness[i] = ns.mindum - ns.deltadum
					} else {
						ns.adjust(config, frontindex)
					}
				}
			}
		}
		ns.share(config)
		ns.minimumdum(config)
		frontindex++
		for i := 0; i < config.Population.Len(); i++ {
			if ns.flag[i] == 2 {
				ns.flag[i] = 3
			} else if ns.flag[i] == 1 {
				ns.flag[i] = 0
			}
		}
	}
	ns.preselect(config)
}

func (ns *NsgaSelection) selection(config *moea.Config, _ [][]float64) int {
	jpick := int(config.RandomNumberGenerator.Float64()) * ns.nremain
	slect := ns.choices[jpick]
	ns.choices[jpick] = ns.choices[ns.nremain]
	ns.nremain--
	return slect
}

func (ns *NsgaSelection) minimumdum(config *moea.Config) {
	/* finding the minimum dummy fitness in the current front */
	ns.mindum = 1000000000.0
	for i := 0; i < config.Population.Len(); i++ {
		if ns.flag[i] == 2 && ns.dumfitness[i] < ns.mindum {
			ns.mindum = ns.dumfitness[i]
		}
	}
}

func (ns *NsgaSelection) adjust(config *moea.Config, index int) {
	/* jack up the fitness of all solutions assigned in a front
	to accomodate remaining ones */
	diff := 2.0*ns.deltadum - ns.mindum
	for i := 0; i < config.Population.Len(); i++ {
		if ns.flag[i] == 1 || ns.flag[i] == 0 {
			continue
		} else {
			ns.dumfitness[i] += diff
		}
	}
	ns.minimumdum(config)
}

func (ns *NsgaSelection) share(config *moea.Config) {
	for i := 0; i < config.Population.Len(); i++ {
		nichecount := 1.0
		if ns.flag[i] == 2 {
			for j := 0; j < config.Population.Len(); j++ {
				if i == j {
					continue
				}
				if ns.flag[j] == 2 {
					d := ns.distance(i, j)
					if d < 0.0 {
						d = (-1.0) * d
					}
					if d <= 0.000001 {
						nichecount++
					} else if d < ns.Dshare {
						nichecount += (1.0 - (d / ns.Dshare)) * (1.0 - (d / ns.Dshare))
					}
				}
			}
		}
		ns.dumfitness[i] /= nichecount
	}
}

func (ns *NsgaSelection) distance(i1, i2 int) float64 {
	sum := 0.0
	v1 := ns.Variables(i1)
	v2 := ns.Variables(i2)
	for i := 0; i < len(v1); i++ {
		sum += math.Sqrt(v1[i]-v2[i])/ns.UpperBounds[i] - ns.LowerBounds[i]
	}
	return math.Sqrt(sum)
}

func (ns *NsgaSelection) preselect(config *moea.Config) {
	sum := 0.0
	for i := 0; i < config.Population.Len(); i++ {
		sum += ns.dumfitness[i]
	}
	dumavg := sum / float64(config.Population.Len())
	if dumavg == 0 {
		for i := 0; i < config.Population.Len(); i++ {
			ns.choices[i] = i
		}
	} else {
		i := 0
		j := 0
		for {
			expected := ns.dumfitness[i] / dumavg
			jassign := int(expected)
			ns.fraction[i] = expected - float64(jassign)
			for jassign > 0 {
				jassign--
				ns.choices[j] = i
				j++
			}
			i++
			if i >= config.Population.Len() {
				break
			}
		}
		i = 0
		for j < config.Population.Len() {
			if i >= config.Population.Len() {
				i = 0
			}
			if ns.fraction[i] > 0.0 && config.RandomNumberGenerator.Flip(ns.fraction[i]) {
				ns.choices[j] = i
				ns.fraction[i] -= 1.0
				j++
			}
			i++
		}
	}
	ns.nremain = config.Population.Len() - 1
}
