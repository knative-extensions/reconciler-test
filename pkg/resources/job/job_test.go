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

package job_test

import (
	"embed"
	"os"

	v1 "k8s.io/api/core/v1"
	"knative.dev/reconciler-test/pkg/manifest"
	"knative.dev/reconciler-test/pkg/resources/job"

	testlog "knative.dev/reconciler-test/pkg/logging"
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
	// apiVersion: batch/v1
	// kind: Job
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: job-container
	//         image: baz
}

func Example_full() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
	}

	opts := []manifest.CfgFn{
		job.WithLabels(map[string]string{
			"color": "green",
		}),
		job.WithAnnotations(map[string]interface{}{
			"app.kubernetes.io/name": "app",
		}),
		job.WithPodLabels(map[string]string{
			"app": "my-app",
		}),
		job.WithPodAnnotations(map[string]interface{}{
			"pod-annotation": "foo",
		}),
		job.WithRestartPolicy(v1.RestartPolicyNever),
		job.WithImagePullPolicy(v1.PullNever),
		job.WithBackoffLimit(20),
		job.WithEnvs(map[string]string{
			"VAR": "VAL",
		}),
		job.WithTTLSecondsAfterFinished(30),
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
	// apiVersion: batch/v1
	// kind: Job
	// metadata:
	//   name: foo
	//   namespace: bar
	//   annotations:
	//     app.kubernetes.io/name: "app"
	//   labels:
	//     color: "green"
	// spec:
	//   backoffLimit: 20
	//   ttlSecondsAfterFinished: 30
	//   template:
	//     metadata:
	//       annotations:
	//         pod-annotation: "foo"
	//       labels:
	//         app: "my-app"
	//     spec:
	//       restartPolicy: Never
	//       containers:
	//       - name: job-container
	//         image: baz
	//         env:
	//         - name: "VAR"
	//           value: "VAL"
	//         imagePullPolicy: Never
}

func Example_withEnvs() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
	}

	job.WithEnvs(map[string]string{
		"VAR1": "VALUE1",
		"VAR2": "VALUE2",
	})(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: batch/v1
	// kind: Job
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: job-container
	//         image: baz
	//         env:
	//         - name: "VAR1"
	//           value: "VALUE1"
	//         - name: "VAR2"
	//           value: "VALUE2"
}

func Example_withPodAnnotations() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
	}

	job.WithPodAnnotations(map[string]interface{}{"app.kubernetes.io/name": "app1"})(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: batch/v1
	// kind: Job
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   template:
	//     metadata:
	//       annotations:
	//         app.kubernetes.io/name: "app1"
	//     spec:
	//       containers:
	//       - name: job-container
	//         image: baz
}

func Example_withPodLabels() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
	}

	job.WithPodLabels(map[string]string{
		"color":   "blue",
		"version": "3",
	})(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: batch/v1
	// kind: Job
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   template:
	//     metadata:
	//       labels:
	//         color: "blue"
	//         version: "3"
	//     spec:
	//       containers:
	//       - name: job-container
	//         image: baz
}

func Example_withImagePullPolicy() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
	}

	job.WithImagePullPolicy(v1.PullAlways)(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: batch/v1
	// kind: Job
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: job-container
	//         image: baz
	//         imagePullPolicy: Always
}

func Example_withRestartPolicy() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
	}

	job.WithRestartPolicy(v1.RestartPolicyAlways)(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: batch/v1
	// kind: Job
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   template:
	//     spec:
	//       restartPolicy: Always
	//       containers:
	//       - name: job-container
	//         image: baz
}

func Example_withBackoffLimit() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
	}

	job.WithBackoffLimit(165)(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: batch/v1
	// kind: Job
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   backoffLimit: 165
	//   template:
	//     spec:
	//       containers:
	//       - name: job-container
	//         image: baz
}

func Example_withTTLSecondsAfterFinished() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"image":     "baz",
	}

	job.WithTTLSecondsAfterFinished(165)(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: batch/v1
	// kind: Job
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   ttlSecondsAfterFinished: 165
	//   template:
	//     spec:
	//       containers:
	//       - name: job-container
	//         image: baz
}
