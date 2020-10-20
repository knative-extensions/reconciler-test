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
	"fmt"
	"os"
	"testing"
	"text/template"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"knative.dev/pkg/injection"
	_ "knative.dev/pkg/system/testing"
	"knative.dev/reconciler-test/rigging/pkg/installer"
	"knative.dev/reconciler-test/rigging/pkg/lifecycle"
)

func TestMain(m *testing.M) {

	fmt.Println("TestMain")

	ctx, _ := injection.EnableInjectionOrDie(nil, nil)

	lifecycle.InjectClients(ctx)

	os.Exit(m.Run())
}

// This test is more for debugging the ko publish process.
func TestKoPublish(t *testing.T) {
	ic, err := installer.ProduceImages()
	if err != nil {
		t.Fatalf("failed to produce images, %s", err)
	}

	templateString := `
	rigging.WithImages(map[string]string{
		{{ range $key, $value := . }}"{{ $key }}": "{{ $value }}",{{ end }}
	}),`

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

// TestEcho is an example simple test.
func TestEcho(t *testing.T) {
	EchoTestImpl(t)
}

// TestBed is an example testbed test.
func TestBed(t *testing.T) {
	t.Skip("test bed not implemented.")
	BedTestImpl(t)
}

//func TestDiff(t *testing.T) {
//	org := map[string]string{
//		"foo": "bar",
//		"baz": "baf",
//	}
//
//	now := map[string]string{
//		"foo": "bar",
//		"baf": "baz",
//		"baz": "boo",
//	}
//
//	if diff := cmp.Diff(org, now); diff != "" {
//		t.Log("FYI, diff on", diff)
//	} else {
//		t.Log("org or now are the same.")
//	}
//
//}
