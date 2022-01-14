//go:build e2e
// +build e2e

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

package eventshub_test

import (
	"testing"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/k8s"
	"knative.dev/reconciler-test/pkg/knative"
	"knative.dev/reconciler-test/test/e2e/eventshub"
)

func TestEventsHubReceiverReadiness(t *testing.T) {
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

	env.Test(ctx, t, eventshub.ReceiverReadiness())
}
