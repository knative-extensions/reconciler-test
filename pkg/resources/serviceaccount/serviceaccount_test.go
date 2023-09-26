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

package serviceaccount_test

import (
	"embed"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testlog "knative.dev/reconciler-test/pkg/logging"
	"knative.dev/reconciler-test/pkg/manifest"
	"knative.dev/reconciler-test/pkg/resources/serviceaccount"
)

//go:embed *.yaml
var yaml embed.FS

func Example_min() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
	}

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: ServiceAccount
	// metadata:
	//   name: foo
	//   namespace: bar
}

func Example_full() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
	}

	controller := true
	opts := []manifest.CfgFn{
		serviceaccount.WithLabels(map[string]string{
			"color": "green",
		}),
		serviceaccount.WithAnnotations(map[string]interface{}{
			"app.kubernetes.io/name": "app",
		}),
		serviceaccount.WithOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion: "eventing.knative.dev/v1",
				Kind:       "Trigger",
				Name:       "my-trigger",
				UID:        "my-trigger-uid",
				Controller: &controller,
			},
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
	// kind: ServiceAccount
	// metadata:
	//   name: foo
	//   namespace: bar
	//   annotations:
	//     app.kubernetes.io/name: "app"
	//   labels:
	//     color: "green"
	//   ownerReferences:
	//     - apiVersion: eventing.knative.dev/v1
	//       kind: Trigger
	//       name: my-trigger
	//       uid: my-trigger-uid
	//       controller: true
}
