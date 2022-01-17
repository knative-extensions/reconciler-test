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
	"flag"
	"os"
	"testing"

	// See: https://github.com/kubernetes/client-go/issues/242
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"knative.dev/pkg/injection"
	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/logging"
)

// global is the singleton instance of GlobalEnvironment. It is used to parse
// the testing config for the test run. The config will specify the cluster
// config as well as the parsing level and state flags.
//goland:noinspection GoUnusedGlobalVariable
var global environment.GlobalEnvironment

// TestMain is the first entry point for `go test`.
func TestMain(m *testing.M) {
	// environment.InitFlags registers state, level and feature filter flags.
	environment.InitFlags(flag.CommandLine)

	// We get a chance to parse flags to include the framework flags for the
	// framework as well as any additional flags included in the integration.
	flag.Parse()

	// EnableInjectionOrDie will enable client injection, this is used by the
	// testing framework for namespace management, and could be leveraged by
	// features to pull Kubernetes clients or the test environment out of the
	// context passed in the features.
	ctx, startInformers := injection.EnableInjectionOrDie(
		logging.WithTestLogger(nil), nil) // nolint
	startInformers()

	// global is used to make instances of Environments, NewGlobalEnvironment
	// is passing and saving the client injection enabled context for use later.
	global = environment.NewGlobalEnvironment(ctx)

	// Run the tests.
	os.Exit(m.Run())
}
