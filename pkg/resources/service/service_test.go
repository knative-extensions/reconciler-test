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

package service_test

import (
	"embed"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	testlog "knative.dev/reconciler-test/pkg/logging"
	"knative.dev/reconciler-test/pkg/manifest"
	"knative.dev/reconciler-test/pkg/resources/service"
)

//go:embed *.yaml
var templates embed.FS

func Example_min() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"ports": []corev1.ServicePort{{
			Port:       80,
			TargetPort: intstr.FromInt(8080),
		}},
	}

	files, err := manifest.ExecuteYAML(ctx, templates, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   ports:
	//     - protocol: TCP
	//       port: 80
	//       targetPort: 8080
}

func Example_full() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
	}

	opts := []manifest.CfgFn{
		service.WithLabels(map[string]string{
			"color": "green",
		}),
		service.WithAnnotations(map[string]interface{}{
			"app.kubernetes.io/name": "app",
		}),
		service.WithType(corev1.ServiceTypeClusterIP),
		service.WithSelectors(map[string]string{
			"app.kubernetes.io/name": "foobar",
		}),
		service.WithExternalName("my-external.name"),
		service.WithPorts([]corev1.ServicePort{{
			Port:       1234,
			TargetPort: intstr.FromInt(5678),
		}}),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	files, err := manifest.ExecuteYAML(ctx, templates, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: foo
	//   namespace: bar
	//   annotations:
	//     app.kubernetes.io/name: "app"
	//   labels:
	//     color: "green"
	// spec:
	//   selector:
	//     app.kubernetes.io/name: "foobar"
	//   ports:
	//     - protocol: TCP
	//       port: 1234
	//       targetPort: 5678
	//   type: ClusterIP
	//   externalName: my-external.name
}

func Example_withSelectors() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"ports": []corev1.ServicePort{{
			Port:       1234,
			TargetPort: intstr.FromInt(5678),
		}},
	}

	service.WithSelectors(map[string]string{
		"color":   "blue",
		"version": "3",
	})(cfg)

	files, err := manifest.ExecuteYAML(ctx, templates, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   selector:
	//     color: "blue"
	//     version: "3"
	//   ports:
	//     - protocol: TCP
	//       port: 1234
	//       targetPort: 5678
}

func Example_withType() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"ports": []corev1.ServicePort{{
			Port:       1234,
			TargetPort: intstr.FromInt(5678),
		}},
	}

	service.WithType(corev1.ServiceTypeLoadBalancer)(cfg)

	files, err := manifest.ExecuteYAML(ctx, templates, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   ports:
	//     - protocol: TCP
	//       port: 1234
	//       targetPort: 5678
	//   type: LoadBalancer
}

func Example_withExternalName() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
		"ports": []corev1.ServicePort{{
			Port:       1234,
			TargetPort: intstr.FromInt(5678),
		}},
	}

	service.WithExternalName("foo.bar")(cfg)

	files, err := manifest.ExecuteYAML(ctx, templates, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   ports:
	//     - protocol: TCP
	//       port: 1234
	//       targetPort: 5678
	//   externalName: foo.bar
}

func Example_withPorts() {
	ctx := testlog.NewContext()
	images := map[string]string{}
	cfg := map[string]interface{}{
		"name":      "foo",
		"namespace": "bar",
	}

	service.WithPorts([]corev1.ServicePort{{
		Port:       1234,
		TargetPort: intstr.FromInt(5678),
	}})(cfg)

	files, err := manifest.ExecuteYAML(ctx, templates, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: foo
	//   namespace: bar
	// spec:
	//   ports:
	//     - protocol: TCP
	//       port: 1234
	//       targetPort: 5678
}
