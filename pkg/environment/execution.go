package environment

import "knative.dev/reconciler-test/pkg/feature"

// reorderSteps reorders the steps based on their timings: Setup, Requirement, Assert, Teardown
func reorderSteps(steps []feature.Step) []feature.Step {
	res := make([]feature.Step, 0, len(steps))

	res = append(res, filterStepTimings(steps, feature.Setup)...)
	res = append(res, filterStepTimings(steps, feature.Requirement)...)
	res = append(res, filterStepTimings(steps, feature.Assert)...)
	res = append(res, filterStepTimings(steps, feature.Teardown)...)

	return res
}

func filterStepTimings(steps []feature.Step, timing feature.Timing) []feature.Step {
	var res []feature.Step
	for _, s := range steps {
		if s.T == timing {
			res = append(res, s)
		}
	}
	return res
}
