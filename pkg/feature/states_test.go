package feature_test

import (
	"testing"

	"knative.dev/reconciler-test/pkg/feature"
)

/*
Copyright 2022 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

func TestState(t *testing.T) {
	t.Parallel()

	for _, tc := range stateTestCases() {
		tc := tc
		t.Run(tc.name(), tc.run)
	}
}

func (tc stateTestCase) run(t *testing.T) {
	t.Parallel()
	got := tc.state.Valid()
	if tc.invalid == got {
		t.Errorf("want %t, got %t", !tc.invalid, got)
	}
}

// stateTestCases returns all possible bitwise OR combinations of
// feature.States: valid ones, invalid, and empty. For example: `Alpha|Stable`,
// `Alpha|Beta|Stable`, etc.
func stateTestCases() []stateTestCase {
	states := []feature.States{
		feature.Alpha,
		feature.Beta,
		feature.Stable,
	}
	states = stateBitwiseOr(statePowerset(states))
	empty := asStateTestCases(false, feature.States(0))
	valid := asStateTestCases(true, states...)
	invalid := asStateTestCases(false,
		feature.States(126), feature.States(127), feature.States(128))

	cases := make([]stateTestCase, 0, len(empty)+len(valid)+len(invalid))
	cases = append(cases, empty...)
	cases = append(cases, valid...)
	cases = append(cases, invalid...)
	return cases
}

// statePowerset returns all combinations for a given array.
func statePowerset(set []feature.States) (subsets [][]feature.States) {
	length := uint(len(set))

	// Go through all possible combinations of objects
	// from 1 (only first object in subset) to 2^length (all objects in subset)
	for subsetBits := 1; subsetBits < (1 << length); subsetBits++ {
		var subset []feature.States

		for object := uint(0); object < length; object++ {
			// checks if object is contained in subset
			// by checking if bit 'object' is set in subsetBits
			if (subsetBits>>object)&1 == 1 {
				// add object to subset
				subset = append(subset, set[object])
			}
		}
		// add subset to subsets
		subsets = append(subsets, subset)
	}
	return subsets
}

func stateBitwiseOr(subsets [][]feature.States) []feature.States {
	levels := make([]feature.States, len(subsets))
	for i, subset := range subsets {
		val := 0
		for _, state := range subset {
			val |= int(state)
		}
		levels[i] = feature.States(val)
	}
	return levels
}

func asStateTestCases(valid bool, states ...feature.States) []stateTestCase {
	cases := make([]stateTestCase, len(states))
	for i, state := range states {
		cases[i] = stateTestCase{
			state: state, invalid: !valid,
		}
	}
	return cases
}

type stateTestCase struct {
	state   feature.States
	invalid bool
}

func (tc stateTestCase) name() string {
	return tc.state.String()
}
