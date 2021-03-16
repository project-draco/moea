package moea

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParettoFrontier(t *testing.T) {
	for _, test := range []struct {
		name   string
		input  Result
		output []IndividualResult
	}{
		{
			name:   "just one individual",
			input:  Result{Individuals: []IndividualResult{{Values: []interface{}{"x"}}}},
			output: []IndividualResult{{Values: []interface{}{"x"}}},
		},
		{
			name: "two individuals",
			input: Result{Individuals: []IndividualResult{
				{Values: []interface{}{"x"}, Objective: []float64{1, 2}},
				{Values: []interface{}{"y"}, Objective: []float64{2, 1}},
			}},
			output: []IndividualResult{
				{Values: []interface{}{"x"}, Objective: []float64{1, 2}},
				{Values: []interface{}{"y"}, Objective: []float64{2, 1}},
			},
		},
		{
			name: "two individuals, one dominating the other",
			input: Result{Individuals: []IndividualResult{
				{Values: []interface{}{"x"}, Objective: []float64{1, 2}},
				{Values: []interface{}{"y"}, Objective: []float64{2, 3}},
			}},
			output: []IndividualResult{
				{Values: []interface{}{"x"}, Objective: []float64{1, 2}},
			},
		},
		{
			name: "three individuals, one dominating the other two",
			input: Result{Individuals: []IndividualResult{
				{Values: []interface{}{"x"}, Objective: []float64{0, 0}},
				{Values: []interface{}{"x"}, Objective: []float64{1, 2}},
				{Values: []interface{}{"y"}, Objective: []float64{2, 1}},
			}},
			output: []IndividualResult{
				{Values: []interface{}{"x"}, Objective: []float64{0, 0}},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if diff := cmp.Diff(test.output, test.input.ParettoFrontier()); diff != "" {
				t.Errorf("diff: %v", diff)
			}
		})
	}
}
