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

package cronjob_test

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
		"labels": map[string]string{
			"app": "foo",
		},
	}

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: batch/v1
	// kind: CronJob
	// metadata:
	//   name: foo
	//   namespace: bar
	//   labels:
	//     app: "foo"
	// spec:
	//   schedule: "* * * * *"
	//   jobTemplate:
	//     metadata:
	//       labels:
	//         app: "foo"
	//     spec:
	//       template:
	//         spec:
	//           containers:
	//             - name: user-container
	//               image: baz
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
			"app":   "foo",
			"bar":   "true",
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
	// kind: CronJob
	// metadata:
	//   name: foo
	//   namespace: bar
	//   annotations:
	//     app.kubernetes.io/name: "app"
	//   labels:
	//     app: "foo"
	//     bar: "true"
	//     color: "green"
	// spec:
	//   schedule: "* * * * *"
	//   jobTemplate:
	//     metadata:
	//       labels:
	//         app: "foo"
	//         bar: "true"
	//         color: "green"
	//     spec:
	//       backoffLimit: 20
	//       ttlSecondsAfterFinished: 30
	//       template:
	//         metadata:
	//           annotations:
	//             pod-annotation: "foo"
	//           labels:
	//             app: my-app
	//         spec:
	//           restartPolicy: Never
	//           containers:
	//             - name: user-container
	//               image: baz
	//               env:
	//               - name: "VAR"
	//                 value: "VAL"
	//               imagePullPolicy: Never
}
