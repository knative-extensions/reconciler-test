/*
 * Copyright 2023 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package deployment_test

import (
	"embed"
	"os"

	v1 "k8s.io/api/core/v1"
	testlog "knative.dev/reconciler-test/pkg/logging"
	"knative.dev/reconciler-test/pkg/manifest"
	"knative.dev/reconciler-test/pkg/resources/deployment"
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
		"selectors": map[string]string{"app": "foo"},
	}

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)

	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   selector:
	//     matchLabels:
	//       app: foo
	//   template:
	//     metadata:
	//       labels:
	//         app: foo
	//     spec:
	//       containers:
	//       - name: user-container
	//         image: baz
}

func Example_full() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
		"selectors": map[string]string{"app": "foo"},
	}

	opts := []manifest.CfgFn{
		deployment.WithLabels(map[string]string{
			"color": "green",
		}),
		deployment.WithAnnotations(map[string]interface{}{
			"app.kubernetes.io/name": "app",
		}),
		deployment.WithPodAnnotations(map[string]interface{}{
			"pod-annotation": "foo",
		}),
		deployment.WithReplicas(6),
		deployment.WithImagePullPolicy(v1.PullNever),
		deployment.WithEnvs(map[string]string{
			"VAR": "VAL",
		}),
		deployment.WithCommand([]string{"sh"}),
		deployment.WithArgs([]string{"-c", "echo \"Hello, Kubernetes!\""}),
		deployment.WithPort(8080),
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
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: bar
	//   annotations:
	//     app.kubernetes.io/name: "app"
	//   labels:
	//     color: green
	// spec:
	//   replicas: 6
	//   selector:
	//     matchLabels:
	//       app: foo
	//   template:
	//     metadata:
	//       annotations:
	//         pod-annotation: "foo"
	//       labels:
	//         app: foo
	//     spec:
	//       containers:
	//       - name: user-container
	//         image: baz
	//         command:
	//         - "sh"
	//         args:
	//         - "-c"
	//         - "echo \"Hello, Kubernetes!\""
	//         ports:
	//         - containerPort: 8080
	//         env:
	//         - name: "VAR"
	//           value: "VAL"
	//         imagePullPolicy: Never
}

func Example_withPodAnnotations() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
		"selectors": map[string]string{"app": "foo"},
	}

	deployment.WithPodAnnotations(map[string]interface{}{
		"pod-annotation": "foo",
	})(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)

	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   selector:
	//     matchLabels:
	//       app: foo
	//   template:
	//     metadata:
	//       annotations:
	//         pod-annotation: "foo"
	//       labels:
	//         app: foo
	//     spec:
	//       containers:
	//       - name: user-container
	//         image: baz
}

func Example_withReplicas() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
		"selectors": map[string]string{"app": "foo"},
	}
	deployment.WithReplicas(6)(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)

	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   replicas: 6
	//   selector:
	//     matchLabels:
	//       app: foo
	//   template:
	//     metadata:
	//       labels:
	//         app: foo
	//     spec:
	//       containers:
	//       - name: user-container
	//         image: baz
}

func Example_withPullPolicy() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
		"selectors": map[string]string{"app": "foo"},
	}
	deployment.WithImagePullPolicy(v1.PullNever)(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)

	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   selector:
	//     matchLabels:
	//       app: foo
	//   template:
	//     metadata:
	//       labels:
	//         app: foo
	//     spec:
	//       containers:
	//       - name: user-container
	//         image: baz
	//         imagePullPolicy: Never
}

func Example_withEnv() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
		"selectors": map[string]string{"app": "foo"},
	}
	deployment.WithEnvs(map[string]string{
		"VAR": "VAL",
	})(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)

	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   selector:
	//     matchLabels:
	//       app: foo
	//   template:
	//     metadata:
	//       labels:
	//         app: foo
	//     spec:
	//       containers:
	//       - name: user-container
	//         image: baz
	//         env:
	//         - name: "VAR"
	//           value: "VAL"
}

func Example_withCommand() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
		"selectors": map[string]string{"app": "foo"},
	}

	deployment.WithCommand([]string{"sh", "-c", "echo \"Hello, Kubernetes!\""})(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)

	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   selector:
	//     matchLabels:
	//       app: foo
	//   template:
	//     metadata:
	//       labels:
	//         app: foo
	//     spec:
	//       containers:
	//       - name: user-container
	//         image: baz
	//         command:
	//         - "sh"
	//         - "-c"
	//         - "echo \"Hello, Kubernetes!\""
}

func Example_withArgs() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
		"selectors": map[string]string{"app": "foo"},
	}

	deployment.WithArgs([]string{"-c", "echo \"Hello, Kubernetes!\""})(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)

	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   selector:
	//     matchLabels:
	//       app: foo
	//   template:
	//     metadata:
	//       labels:
	//         app: foo
	//     spec:
	//       containers:
	//       - name: user-container
	//         image: baz
	//         args:
	//         - "-c"
	//         - "echo \"Hello, Kubernetes!\""
}

func Example_withPort() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
		"selectors": map[string]string{"app": "foo"},
	}

	deployment.WithPort(8080)(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)

	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   selector:
	//     matchLabels:
	//       app: foo
	//   template:
	//     metadata:
	//       labels:
	//         app: foo
	//     spec:
	//       containers:
	//       - name: user-container
	//         image: baz
	//         ports:
	//         - containerPort: 8080
}
