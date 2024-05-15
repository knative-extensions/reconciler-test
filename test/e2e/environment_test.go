//go:build e2e
// +build e2e

/*
Copyright 2020 The Knative Authors

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

package e2e

import (
	"context"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/feature"
)

func TestTimingConstraints(t *testing.T) {
	testTimingConstraints(t, false)
}

func TestTimingConstraintsParallel(t *testing.T) {
	testTimingConstraints(t, true)
}

func testTimingConstraints(t *testing.T, isParallel bool) {
	// Signal to the go test framework that this test can be run in parallel
	// with other tests.
	t.Parallel()

	ctx, env := global.Environment(environment.Managed(t))

	// Build the feature
	feat := feature.NewFeature()

	setupCounter := int32(0)
	assertCounter := int32(0)
	requirementCounter := int32(0)
	teardownCounter := int32(0)

	setupGroupCounter := int32(0)
	requirementGroupCounter := int32(0)
	assertGroupCounter := int32(0)
	teardownGroupCounter := int32(0)

	feat.Group("my-group", func(f *feature.Feature) {
		f.Setup("setup group", func(ctx context.Context, t feature.T) {
			atomic.AddInt32(&setupGroupCounter, 1)
			// group
			verifyCounter(&requirementGroupCounter, 0, t)
			verifyCounter(&assertGroupCounter, 0, t)
			verifyCounter(&teardownGroupCounter, 0, t)

			// feature
			verifyCounter(&setupCounter, 0, t)
			verifyCounter(&requirementCounter, 0, t)
			verifyCounter(&assertCounter, 0, t)
			verifyCounter(&teardownCounter, 0, t)
		})
		f.Requirement("requirement group", func(ctx context.Context, t feature.T) {
			atomic.AddInt32(&requirementGroupCounter, 1)
			// group
			verifyCounter(&setupGroupCounter, 1, t)
			verifyCounter(&assertGroupCounter, 0, t)
			verifyCounter(&teardownGroupCounter, 0, t)

			// feature
			verifyCounter(&setupCounter, 0, t)
			verifyCounter(&requirementCounter, 0, t)
			verifyCounter(&assertCounter, 0, t)
			verifyCounter(&teardownCounter, 0, t)
		})
		f.Assert("assert group", func(ctx context.Context, t feature.T) {
			atomic.AddInt32(&assertGroupCounter, 1)
			// group
			verifyCounter(&setupGroupCounter, 1, t)
			verifyCounter(&requirementGroupCounter, 1, t)
			verifyCounter(&teardownGroupCounter, 0, t)

			// feature
			verifyCounter(&setupCounter, 0, t)
			verifyCounter(&requirementCounter, 0, t)
			verifyCounter(&assertCounter, 0, t)
			verifyCounter(&teardownCounter, 0, t)
		})
		f.Teardown("teardown group", func(ctx context.Context, t feature.T) {
			atomic.AddInt32(&teardownGroupCounter, 1)
			// group
			verifyCounter(&setupGroupCounter, 1, t)
			verifyCounter(&requirementGroupCounter, 1, t)
			verifyCounter(&assertGroupCounter, 1, t)

			// feature
			verifyCounter(&setupCounter, 0, t)
			verifyCounter(&requirementCounter, 0, t)
			verifyCounter(&assertCounter, 0, t)
			verifyCounter(&teardownCounter, 0, t)
		})
	})

	incrementSetupCounter := func(ctx context.Context, t feature.T) {
		atomic.AddInt32(&setupCounter, 1)
		verifyCounter(&requirementCounter, 0, t)
		verifyCounter(&assertCounter, 0, t)
		verifyCounter(&teardownCounter, 0, t)

		// group
		verifyCounter(&setupGroupCounter, 1, t)
		verifyCounter(&requirementGroupCounter, 1, t)
		verifyCounter(&assertGroupCounter, 1, t)
		verifyCounter(&teardownGroupCounter, 1, t)
	}
	incrementRequirementCounter := func(ctx context.Context, t feature.T) {
		verifyCounter(&setupCounter, 3, t)
		atomic.AddInt32(&requirementCounter, 1)
		verifyCounter(&assertCounter, 0, t)
		verifyCounter(&teardownCounter, 0, t)

		// group
		verifyCounter(&setupGroupCounter, 1, t)
		verifyCounter(&requirementGroupCounter, 1, t)
		verifyCounter(&assertGroupCounter, 1, t)
		verifyCounter(&teardownGroupCounter, 1, t)
	}

	incrementAssertCounter := func(ctx context.Context, t feature.T) {
		verifyCounter(&requirementCounter, 5, t)
		atomic.AddInt32(&assertCounter, 1)
		verifyCounter(&teardownCounter, 0, t)

		// group
		verifyCounter(&setupGroupCounter, 1, t)
		verifyCounter(&requirementGroupCounter, 1, t)
		verifyCounter(&assertGroupCounter, 1, t)
		verifyCounter(&teardownGroupCounter, 1, t)
	}

	incrementTeardownCounter := func(ctx context.Context, t feature.T) {
		verifyCounter(&assertCounter, 4, t)
		atomic.AddInt32(&teardownCounter, 1)
		verifyCounter(&setupCounter, 3, t)
		verifyCounter(&requirementCounter, 5, t)

		// group
		verifyCounter(&setupGroupCounter, 1, t)
		verifyCounter(&requirementGroupCounter, 1, t)
		verifyCounter(&assertGroupCounter, 1, t)
		verifyCounter(&teardownGroupCounter, 1, t)
	}

	feat.Setup("setup1", incrementSetupCounter)
	feat.Setup("setup2", incrementSetupCounter)
	feat.Setup("setup3", incrementSetupCounter)
	feat.Requirement("requirement1", incrementRequirementCounter)
	feat.Requirement("requirement2", incrementRequirementCounter)
	feat.Requirement("requirement3", incrementRequirementCounter)
	feat.Requirement("requirement4", incrementRequirementCounter)
	feat.Requirement("requirement5", incrementRequirementCounter)
	feat.Stable("A cool feature").
		Must("aaa", incrementAssertCounter).
		Must("bbb", incrementAssertCounter).
		Must("ccc", incrementAssertCounter).
		Must("ddd", incrementAssertCounter)
	feat.Teardown("teardown0", incrementTeardownCounter)
	feat.Teardown("teardown1", incrementTeardownCounter)

	if isParallel {
		// This subtest makes sure that the parallel tests ends before running other tests like
		// "Verify teardown counter"
		t.Run("test", func(t *testing.T) {
			env.ParallelTest(ctx, t, feat)
		})
	} else {
		env.Test(ctx, t, feat)
	}

	verifyCounter(&teardownCounter, 2, t)
	verifyCounter(&teardownGroupCounter, 1, t)
}

func verifyCounter(counter *int32, expected int32, t feature.T) {
	got := atomic.LoadInt32(counter)
	if got != expected {
		t.Errorf("expected counter to be %d got %d:\n%s\n", expected, got, string(debug.Stack()))
	}
}

func TestContextLifecycle(t *testing.T) {
	// Signal to the go test framework that this test can be run in parallel
	// with other tests.
	t.Parallel()

	ctx, env := global.Environment(environment.Managed(t))

	// Build the feature
	feat := feature.NewFeature()

	var ctxVal context.Context
	feat.Setup("get the context", func(ctx context.Context, t feature.T) {
		ctxVal = ctx
	})
	feat.Assert("check the context is closed and different", func(ctx context.Context, t feature.T) {
		require.NotSame(t, ctx, ctxVal)
		require.NotNil(t, ctxVal.Err())
		require.Nil(t, ctx.Err())
	})

	env.Test(ctx, t, feat)
}

func appender(stringBuilder *strings.Builder, val string) feature.StepFn {
	return func(ctx context.Context, t feature.T) {
		stringBuilder.WriteString(val)
	}
}

func TestPrerequisiteShouldRunFalse(t *testing.T) {
	t.Parallel()

	ctx, env := global.Environment(environment.Managed(t))

	f := feature.NewFeature()
	f.Prerequisite("other timings should not run", func(ctx context.Context, t feature.T) (feature.PrerequisiteResult, error) {
		return feature.PrerequisiteResult{ShouldRun: false}, nil
	})

	mustNeverHappen := func(ctx context.Context, t feature.T) {
		t.Fatal("bug: this must never happen")
	}

	f.Setup("failed setup", mustNeverHappen)
	f.Requirement("failed setup", mustNeverHappen)
	f.Teardown("teardown", mustNeverHappen)

	env.Test(ctx, t, f)
}

func TestPrerequisiteShouldRunTrue(t *testing.T) {
	ctx, env := global.Environment(environment.Managed(t))

	f := feature.NewFeatureNamed("should run all timings")
	f.Prerequisite("other timings should run", func(ctx context.Context, t feature.T) (feature.PrerequisiteResult, error) {
		return feature.PrerequisiteResult{ShouldRun: true}, nil
	})

	setupCounter := int32(0)
	requirementCounter := int32(0)
	teardownCounter := int32(0)

	f.Setup("setup", func(ctx context.Context, t feature.T) {
		atomic.AddInt32(&setupCounter, 1)
	})
	f.Requirement("requirement", func(ctx context.Context, t feature.T) {
		verifyCounter(&setupCounter, 1, t)
		atomic.AddInt32(&requirementCounter, 1)
	})
	f.Teardown("teardown", func(ctx context.Context, t feature.T) {
		verifyCounter(&setupCounter, 1, t)
		verifyCounter(&requirementCounter, 1, t)
		atomic.AddInt32(&teardownCounter, 1)
	})

	env.Test(ctx, t, f)
	env.Test(ctx, t, feature.StepFn(func(ctx context.Context, t feature.T) {
		verifyCounter(&setupCounter, 1, t)
		verifyCounter(&requirementCounter, 1, t)
		verifyCounter(&teardownCounter, 1, t)
	}).AsFeature())
}
