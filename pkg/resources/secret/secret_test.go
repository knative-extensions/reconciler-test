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

package secret_test

import (
	"embed"
	"os"

	v1 "k8s.io/api/core/v1"
	"knative.dev/reconciler-test/pkg/manifest"
	"knative.dev/reconciler-test/pkg/resources/secret"

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
	}

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Secret
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

	opts := []manifest.CfgFn{
		secret.WithLabels(map[string]string{
			"color": "green",
		}),
		secret.WithAnnotations(map[string]interface{}{
			"app.kubernetes.io/name": "app",
		}),
		secret.WithType(v1.SecretTypeOpaque),
		secret.WithData(map[string][]byte{
			"key1": []byte("val1"),
		}),
		secret.WithStringData(map[string]string{
			"key2": "val2",
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
	// kind: Secret
	// metadata:
	//   name: foo
	//   namespace: bar
	//   annotations:
	//     app.kubernetes.io/name: "app"
	//   labels:
	//     color: green
	// type: Opaque
	// data:
	//   key1: dmFsMQ==
	// stringData:
	//   key2: val2
}

func Example_WithData() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
	}

	secret.WithData(map[string][]byte{
		"color":   []byte("blue"),
		"version": []byte("3"),
	})(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Secret
	// metadata:
	//   name: foo
	//   namespace: bar
	// data:
	//   color: Ymx1ZQ==
	//   version: Mw==
}

func Example_WithStringData() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
	}

	secret.WithStringData(map[string]string{
		"color":   "blue",
		"version": "3",
	})(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Secret
	// metadata:
	//   name: foo
	//   namespace: bar
	// stringData:
	//   color: blue
	//   version: 3
}

func Example_WithType() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
	}

	secret.WithType(v1.SecretTypeDockerConfigJson)(cfg)

	files, err := manifest.ExecuteYAML(ctx, yaml, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Secret
	// metadata:
	//   name: foo
	//   namespace: bar
	// type: kubernetes.io/dockerconfigjson
}
