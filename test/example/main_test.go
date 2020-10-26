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
	"flag"
	"fmt"
	"os"
	"testing"
	"text/template"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"knative.dev/pkg/injection"
	_ "knative.dev/pkg/system/testing"
	"knative.dev/reconciler-test/pkg/environment"
)

var (
	global environment.GlobalEnvironment
)

func TestMain(m *testing.M) {
	ctx, startInformers := injection.EnableInjectionOrDie(nil, nil) //nolint

	global = environment.NewGlobalEnvironment(ctx)
	global.InitFlags(flag.CommandLine)
	flag.Parse()

	startInformers()

	os.Exit(m.Run())
}

// This test is more for debugging the ko publish process.
func TestKoPublish(t *testing.T) {
	fmt.Println("TestKoPublish")
	ic, err := environment.ProduceImages()
	if err != nil {
		panic(fmt.Errorf("failed to produce images, %s", err))
	}

	templateString := `
// The following could be used to bypass the image generation process.

import "knative.dev/reconciler-test/pkg/environment"

func init() {
	environment.WithImages(map[string]string{
		{{ range $key, $value := . }}"{{ $key }}": "{{ $value }}",
		{{ end }}
	})
}
`

	tp := template.New("t")
	temp, err := tp.Parse(templateString)
	if err != nil {
		panic(err)
	}

	err = temp.Execute(os.Stdout, ic)
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Fprint(os.Stdout, "\n\n")
}

// Rest of e2e tests go below:
//
// TestEcho is an example simple test.
func TestEcho(t *testing.T) {
	t.Parallel()

	// Create an environment to run the tests in from the global environment.
	ctx, env := global.Environment()

	f := EchoFeature()

	// Now is the chance to modify the feature to add additional preconditions or assertions.

	env.Test(ctx, t, f)

	// Calling finish on the environment cleans it up and removes the namespace.
	env.Finish()
}

// TestRecorder is an example simple test.
func TestRecorder(t *testing.T) {
	t.Parallel()

	// Create an environment to run the tests in from the global environment.
	ctx, env := global.Environment()

	// Now is the chance to inject extra things into context.

	f := RecorderFeature()

	// Now is the chance to modify the feature to add additional preconditions or assertions.

	env.Test(ctx, t, f)

	// Calling finish on the environment cleans it up and removes the namespace.
	env.Finish()
}
