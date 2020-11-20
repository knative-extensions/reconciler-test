package feature

import (
	"context"
	"testing"
)

// CollapseSteps is used by the framework to impose the execution constraints of the different steps
// Steps with Setup, Requirement and Teardown timings are run sequentially in order
// Steps with Assert timing are run in parallel
func CollapseSteps(steps []Step) []Step {
	var setup *Step
	var requirement *Step
	var asserts []Step
	var teardown *Step

	for i, s := range steps {
		if s.T == Setup {
			setup = composeStep(setup, &steps[i])
		} else if s.T == Requirement {
			requirement = composeStep(requirement, &steps[i])
		} else if s.T == Assert {
			asserts = append(asserts, parallelizeStep(s))
		} else if s.T == Teardown {
			teardown = composeStep(teardown, &steps[i])
		}
	}

	var result []Step
	if setup != nil {
		result = append(result, *setup)
	}
	if requirement != nil {
		result = append(result, *requirement)
	}
	result = append(result, asserts...)
	if teardown != nil {
		result = append(result, *teardown)
	}

	return result
}

func composeStep(x *Step, y *Step) *Step {
	if x == nil {
		return y
	}
	if y == nil {
		return x
	}
	return &Step{
		Name: x.Name + " and " + y.Name,
		S:    x.S,
		L:    x.L,
		T:    x.T,
		Fn: func(ctx context.Context, t *testing.T) {
			t.Helper()
			x.Fn(ctx, t)
			y.Fn(ctx, t)
		},
	}
}

func parallelizeStep(x Step) Step {
	return Step{
		Name: x.Name,
		S:    x.S,
		L:    x.L,
		T:    x.T,
		Fn: func(ctx context.Context, t *testing.T) {
			t.Parallel()
			t.Helper()
			x.Fn(ctx, t)
		},
	}
}
