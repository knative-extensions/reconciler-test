/*
Copyright 2021 The Knative Authors

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

package pod_test

import (
	"embed"
	"os"

	v1 "k8s.io/api/core/v1"
	testlog "knative.dev/reconciler-test/pkg/logging"
	"knative.dev/reconciler-test/pkg/manifest"
	"knative.dev/reconciler-test/pkg/resources/pod"
)

//go:embed *.yaml
var yaml embed.FS

func Example_min() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
	}

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Pod
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   containers:
	//   - name: user-container
	//     image: baz
}

func Example_full() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":  "foo",
		"image": "baz",
	}

	opts := []manifest.CfgFn{
		pod.WithLabels(map[string]string{
			"color": "green",
		}),
		pod.WithAnnotations(map[string]interface{}{
			"app.kubernetes.io/name": "app",
		}),
		pod.WithNamespace("bar"),
		pod.WithCommand([]string{"sh"}),
		pod.WithArgs([]string{"-c", "echo \"Hello, Kubernetes!\""}),
		pod.WithImagePullPolicy(v1.PullNever),
		pod.WithEnvs(map[string]string{
			"VAR": "VAL",
		}),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Pod
	// metadata:
	//   name: foo
	//   namespace: bar
	//   annotations:
	//     app.kubernetes.io/name: "app"
	//   labels:
	//     color: green
	// spec:
	//   containers:
	//   - name: user-container
	//     image: baz
	//     command:
	//     - "sh"
	//     args:
	//     - "-c"
	//     - "echo \"Hello, Kubernetes!\""
	//     env:
	//     - name: "VAR"
	//       value: "VAL"
	//     imagePullPolicy: Never
}

func Example_withCommand() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
	}

	pod.WithCommand([]string{"sh", "-c", "echo \"Hello, Kubernetes!\""})(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Pod
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   containers:
	//   - name: user-container
	//     image: baz
	//     command:
	//     - "sh"
	//     - "-c"
	//     - "echo \"Hello, Kubernetes!\""
}

func Example_withArgs() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
	}

	pod.WithArgs([]string{"-c", "echo \"Hello, Kubernetes!\""})(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Pod
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   containers:
	//   - name: user-container
	//     image: baz
	//     args:
	//     - "-c"
	//     - "echo \"Hello, Kubernetes!\""
}

func Example_withNamespace() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
	}

	pod.WithNamespace("new-namespace")(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Pod
	// metadata:
	//   name: foo
	//   namespace: new-namespace
	// spec:
	//   containers:
	//   - name: user-container
	//     image: baz
}
