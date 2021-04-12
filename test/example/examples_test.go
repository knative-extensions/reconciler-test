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

package example

import (
	"testing"
	"time"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	_ "knative.dev/pkg/system/testing"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/k8s"
	"knative.dev/reconciler-test/pkg/knative"
)

// TestRecorder is an example simple test.
func TestRecorder(t *testing.T) {
	// Signal to the go test framework that this test can be run in parallel
	// with other tests.
	t.Parallel()

	// Create an instance of an environment. The environment will be configured
	// with any relevant configuration and settings based on the global
	// environment settings. Additional options can be passed to Environment()
	// if customization is required.
	ctx, env := global.Environment(
		environment.Managed(t), // Will call env.Finish() when the test exits.
		knative.WithKnativeNamespace("knative-reconciler-test"),
		knative.WithLoggingConfig,
		knative.WithTracingConfig,
		k8s.WithEventListener,
	)

	// With the instance of an Environment, perform one or more calls to Test().
	env.Test(ctx, t, RecorderFeature())
	env.Test(ctx, t, RecorderFeatureYAML())
}

// TestProber is an example simple test.
func TestProber(t *testing.T) {
	// Signal to the go test framework that this test can be run in parallel
	// with other tests.
	t.Parallel()

	// Create an instance of an environment. The environment will be configured
	// with any relevant configuration and settings based on the global
	// environment settings. Additional options can be passed to Environment()
	// if customization is required.
	ctx, env := global.Environment(
		environment.Managed(t), // Will call env.Finish() when the test exits.
		knative.WithKnativeNamespace("knative-reconciler-test"),
		knative.WithLoggingConfig,
		knative.WithTracingConfig,
		k8s.WithEventListener,
	)

	// With the instance of an Environment, perform one or more calls to Test().
	env.Test(ctx, t, ProberFeature())
	env.Test(ctx, t, ProberFeatureYAML())
}

// TestEcho is an example simple test.
func TestEcho(t *testing.T) {
	// Signal to the go test framework that this test can be run in parallel
	// with other tests.
	t.Parallel()

	// Create an environment to run the tests in from the global environment.
	ctx, env := global.Environment(
		environment.WithPollTimings(time.Second*5, time.Minute*2), // Override the default poll timings.
		environment.Managed(t)) // Call env.Finish() on test completion.

	f := EchoFeature()

	// Now is the chance to modify the feature to add additional preconditions or assertions.

	env.Test(ctx, t, f)

	// note: we can run other features in this environment if we understand the side-effects.
	// env.Test(ctx, t, SomeOtherFeature())
}

// TestEchoSet is an example simple test set.
func TestEchoSet(t *testing.T) {
	// Signal to the go test framework that this test can be run in parallel
	// with other tests.
	t.Parallel()

	// Create an environment to run the tests in from the global environment.
	ctx, env := global.Environment(environment.Managed(t))

	fs := EchoFeatureSet()

	// Now is the chance to modify the feature to add additional preconditions or assertions.

	env.TestSet(ctx, t, fs)

	// note: we can run other features in this environment if we understand the side-effects.
	// env.Test(ctx, t, SomeOtherFeature())
}
