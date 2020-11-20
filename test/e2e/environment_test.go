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
	"testing"

	"github.com/stretchr/testify/require"

	"knative.dev/reconciler-test/pkg/feature"
)

func TestTimingConstraints(t *testing.T) {
	// Signal to the go test framework that this test can be run in parallel
	// with other tests.
	t.Parallel()

	ctx, env := global.Environment()

	// We assert at the end on this string
	stringBuilder := &strings.Builder{}

	// Build the feature
	feat := &feature.Feature{}

	feat.Setup("setup1", appender(stringBuilder, "setup1"))
	feat.Setup("setup2", appender(stringBuilder, "setup2"))
	feat.Setup("setup3", appender(stringBuilder, "setup3"))
	feat.Requirement("requirement1", appender(stringBuilder, "requirement1"))
	feat.Requirement("requirement2", appender(stringBuilder, "requirement2"))
	feat.Requirement("requirement3", appender(stringBuilder, "requirement3"))
	feat.Stable("aaa").Must("bbb", func(ctx context.Context, t *testing.T) {})
	feat.Teardown("teardown1", appender(stringBuilder, "teardown1"))
	feat.Teardown("teardown2", appender(stringBuilder, "teardown2"))
	feat.Teardown("teardown3", appender(stringBuilder, "teardown3"))

	env.Test(ctx, t, feat)

	require.Equal(t, "setup1setup2setup3requirement1requirement2requirement3teardown1teardown2teardown3", stringBuilder.String())
}

func appender(stringBuilder *strings.Builder, val string) feature.StepFn {
	return func(ctx context.Context, t *testing.T) {
		stringBuilder.WriteString(val)
	}
}
