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
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/feature"
)

func TestTimingConstraints(t *testing.T) {
	// Signal to the go test framework that this test can be run in parallel
	// with other tests.
	t.Parallel()

	ctx, env := global.Environment(environment.Managed(t))

	// Build the feature
	feat := feature.NewFeature()

	setupCounter := int32(0)
	incrementSetupCounter := func(ctx context.Context, t feature.T) {
		atomic.AddInt32(&setupCounter, 1)
	}
	requirementCounter := int32(0)
	incrementRequirementCounter := func(ctx context.Context, t feature.T) {
		verifyCounter(&setupCounter, 3, t)
		atomic.AddInt32(&requirementCounter, 1)
	}

	assertCounter := int32(0)
	incrementAssertCounter := func(ctx context.Context, t feature.T) {
		verifyCounter(&requirementCounter, 5, t)
		atomic.AddInt32(&assertCounter, 1)
	}

	teardownCounter := int32(0)
	incrementTeardownCounter := func(ctx context.Context, t feature.T) {
		verifyCounter(&assertCounter, 4, t)
		atomic.AddInt32(&teardownCounter, 1)
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

	env.Test(ctx, t, feat)

	verifyCounter(&teardownCounter, 2, t)
}

func verifyCounter(counter *int32, expected int32, t feature.T) {
	got := atomic.LoadInt32(counter)
	if got != expected {
		t.Errorf("expected requirement counter to be 4 got %d", got)
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
