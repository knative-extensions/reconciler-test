package feature

// ReorderSteps reorders steps as follows: Setup, Requirement, Assert and Teardown
func ReorderSteps(steps []Step) []Step {
	var (
		setup, requirement, asserts, teardown []Step
	)

	for _, s := range steps {
		switch s.T {
		case Setup:
			setup = append(setup, s)
		case Requirement:
			requirement = append(requirement, s)
		case Assert:
			asserts = append(asserts, s)
		case Teardown:
			teardown = append(teardown, s)
		}
	}

	var result []Step
	result = append(result, setup...)
	result = append(result, requirement...)
	result = append(result, asserts...)
	result = append(result, teardown...)

	return result
}
