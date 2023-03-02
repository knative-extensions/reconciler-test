package feature_test

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

import (
	"testing"

	"knative.dev/reconciler-test/pkg/feature"
)

func TestLevel(t *testing.T) {
	t.Parallel()

	for _, tc := range levelTestCases() {
		tc := tc
		t.Run(tc.name(), tc.run)
	}
}

func (tc levelTestCase) run(t *testing.T) {
	t.Parallel()
	got := tc.level.Valid()
	if tc.invalid == got {
		t.Errorf("want %t, got %t", !tc.invalid, got)
	}
}

// levelTestCases returns all possible bitwise OR combinations of
// feature.Levels: valid ones, invalid, and empty. For example: `Must|Should`,
// `Must|MustNot|Should|ShouldNot|May`, or `Invalid(128)` etc.
func levelTestCases() []levelTestCase {
	levels := []feature.Levels{
		feature.Must,
		feature.MustNot,
		feature.Should,
		feature.ShouldNot,
		feature.May,
	}
	levels = levelBitwiseOr(levelPowerset(levels))
	empty := asLevelTestCases(false, feature.Levels(0))
	valid := asLevelTestCases(true, levels...)
	invalid := asLevelTestCases(false,
		feature.Levels(126), feature.Levels(127), feature.Levels(128))

	cases := make([]levelTestCase, 0, len(empty)+len(valid)+len(invalid))
	cases = append(cases, empty...)
	cases = append(cases, valid...)
	cases = append(cases, invalid...)
	return cases
}

// levelPowerset returns all combinations for a given array.
func levelPowerset(set []feature.Levels) (subsets [][]feature.Levels) {
	length := uint(len(set))

	// Go through all possible combinations of objects
	// from 1 (only first object in subset) to 2^length (all objects in subset)
	for subsetBits := 1; subsetBits < (1 << length); subsetBits++ {
		var subset []feature.Levels

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

func levelBitwiseOr(subsets [][]feature.Levels) []feature.Levels {
	levels := make([]feature.Levels, len(subsets))
	for i, subset := range subsets {
		val := 0
		for _, level := range subset {
			val |= int(level)
		}
		levels[i] = feature.Levels(val)
	}
	return levels
}

func asLevelTestCases(valid bool, levels ...feature.Levels) []levelTestCase {
	cases := make([]levelTestCase, len(levels))
	for i, level := range levels {
		cases[i] = levelTestCase{
			level: level, invalid: !valid,
		}
	}
	return cases
}

type levelTestCase struct {
	level   feature.Levels
	invalid bool
}

func (tc levelTestCase) name() string {
	return tc.level.String()
}
