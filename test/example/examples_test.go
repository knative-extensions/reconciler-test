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

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	_ "knative.dev/pkg/system/testing"

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
		knative.WithKnativeNamespace("knative-reconciler-test"),
		knative.WithLoggingConfig,
		knative.WithTracingConfig,
		k8s.WithEventListener,
	)

	// With the instance of an Environment, perform one or more calls to Test().
	env.Test(ctx, t, RecorderFeature())

	// Call Finish() on the Environment when finished. This will clean up any
	// ephemeral resources created from global.Environment() and env.Test().
	env.Finish()
}

// TestEcho is an example simple test.
func TestEcho(t *testing.T) {
	// Signal to the go test framework that this test can be run in parallel
	// with other tests.
	t.Parallel()

	// Create an environment to run the tests in from the global environment.
	ctx, env := global.Environment()

	f := EchoFeature()

	// Now is the chance to modify the feature to add additional preconditions or assertions.

	env.Test(ctx, t, f)

	// note: we can run other features in this environment if we understand the side-effects.
	// env.Test(ctx, t, SomeOtherFeature())

	// Calling finish on the environment cleans it up and removes the namespace.
	env.Finish()
}

// TestEchoSet is an example simple test set.
func TestEchoSet(t *testing.T) {
	// Signal to the go test framework that this test can be run in parallel
	// with other tests.
	t.Parallel()

	// Create an environment to run the tests in from the global environment.
	ctx, env := global.Environment()

	fs := EchoFeatureSet()

	// Now is the chance to modify the feature to add additional preconditions or assertions.

	env.TestSet(ctx, t, fs)

	// note: we can run other features in this environment if we understand the side-effects.
	// env.Test(ctx, t, SomeOtherFeature())

	// Calling finish on the environment cleans it up and removes the namespace.
	env.Finish()
}
