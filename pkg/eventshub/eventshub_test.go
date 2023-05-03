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

package eventshub_test

import (
	"embed"
	"encoding/json"
	"os"
	"testing"

	"knative.dev/reconciler-test/pkg/eventshub"
	testlog "knative.dev/reconciler-test/pkg/logging"
	"knative.dev/reconciler-test/pkg/manifest"
)

//go:embed 102-service.yaml 103-pod.yaml
var templates embed.FS

//go:embed 104-forwarder.yaml
var templatesForwarder embed.FS

func Example() {
	ctx := testlog.NewContext()
	images := map[string]string{
		"ko://knative.dev/reconciler-test/cmd/eventshub": "uri://a-real-container",
	}
	cfg := map[string]interface{}{
		"name":          "hubhub",
		"serviceName":   "hubhub",
		"namespace":     "example",
		"image":         "ko://knative.dev/reconciler-test/cmd/eventshub",
		"withReadiness": true,
		"envs": map[string]string{
			"foo": "bar",
			"baz": "boof",
		},
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
	//   name: hubhub
	//   namespace: example
	// spec:
	//   selector:
	//     app: eventshub-hubhub
	//   ports:
	//     - protocol: TCP
	//       port: 80
	//       targetPort: 8080
	// ---
	// apiVersion: v1
	// kind: Pod
	// metadata:
	//   name: hubhub
	//   namespace: example
	//   labels:
	//     app: eventshub-hubhub
	// spec:
	//   serviceAccountName: "example"
	//   restartPolicy: "OnFailure"
	//   containers:
	//     - name: eventshub
	//       image: uri://a-real-container
	//       imagePullPolicy: "IfNotPresent"
	//       readinessProbe:
	//         httpGet:
	//           port: 8080
	//           path: /health/ready
	//       env:
	//         - name: "SYSTEM_NAMESPACE"
	//           valueFrom:
	//             fieldRef:
	//               fieldPath: "metadata.namespace"
	//         - name: "POD_NAME"
	//           valueFrom:
	//             fieldRef:
	//               fieldPath: "metadata.name"
	//         - name: "EVENT_LOGS"
	//           value: "recorder,logger"
	//         - name: "baz"
	//           value: "boof"
	//         - name: "foo"
	//           value: "bar"
}

func ExampleIstioAnnotation() {
	ctx := testlog.NewContext()
	images := map[string]string{
		"ko://knative.dev/reconciler-test/cmd/eventshub": "uri://a-real-container",
	}
	cfg := map[string]interface{}{
		"name":          "hubhub",
		"serviceName":   "hubhub",
		"namespace":     "example",
		"image":         "ko://knative.dev/reconciler-test/cmd/eventshub",
		"withReadiness": true,
		"envs": map[string]string{
			"foo": "bar",
			"baz": "boof",
		},
	}

	manifest.WithIstioPodAnnotations(cfg)

	files, err := manifest.ExecuteYAML(ctx, templates, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: hubhub
	//   namespace: example
	// spec:
	//   selector:
	//     app: eventshub-hubhub
	//   ports:
	//     - protocol: TCP
	//       port: 80
	//       targetPort: 8080
	// ---
	// apiVersion: v1
	// kind: Pod
	// metadata:
	//   name: hubhub
	//   namespace: example
	//   labels:
	//     app: eventshub-hubhub
	//   annotations:
	//       sidecar.istio.io/inject: "true"
	//       sidecar.istio.io/rewriteAppHTTPProbers: "true"
	// spec:
	//   serviceAccountName: "example"
	//   restartPolicy: "OnFailure"
	//   containers:
	//     - name: eventshub
	//       image: uri://a-real-container
	//       imagePullPolicy: "IfNotPresent"
	//       readinessProbe:
	//         httpGet:
	//           port: 8080
	//           path: /health/ready
	//       env:
	//         - name: "SYSTEM_NAMESPACE"
	//           valueFrom:
	//             fieldRef:
	//               fieldPath: "metadata.namespace"
	//         - name: "POD_NAME"
	//           valueFrom:
	//             fieldRef:
	//               fieldPath: "metadata.name"
	//         - name: "EVENT_LOGS"
	//           value: "recorder,logger"
	//         - name: "baz"
	//           value: "boof"
	//         - name: "foo"
	//           value: "bar"
}

func ExampleNoReadiness() {
	ctx := testlog.NewContext()
	images := map[string]string{
		"ko://knative.dev/reconciler-test/cmd/eventshub": "uri://a-real-container",
	}
	cfg := map[string]interface{}{
		"name":        "hubhub",
		"serviceName": "hubhub",
		"namespace":   "example",
		"image":       "ko://knative.dev/reconciler-test/cmd/eventshub",
		"envs": map[string]string{
			"foo": "bar",
			"baz": "boof",
		},
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
	//   name: hubhub
	//   namespace: example
	// spec:
	//   selector:
	//     app: eventshub-hubhub
	//   ports:
	//     - protocol: TCP
	//       port: 80
	//       targetPort: 8080
	// ---
	// apiVersion: v1
	// kind: Pod
	// metadata:
	//   name: hubhub
	//   namespace: example
	//   labels:
	//     app: eventshub-hubhub
	// spec:
	//   serviceAccountName: "example"
	//   restartPolicy: "OnFailure"
	//   containers:
	//     - name: eventshub
	//       image: uri://a-real-container
	//       imagePullPolicy: "IfNotPresent"
	//       env:
	//         - name: "SYSTEM_NAMESPACE"
	//           valueFrom:
	//             fieldRef:
	//               fieldPath: "metadata.namespace"
	//         - name: "POD_NAME"
	//           valueFrom:
	//             fieldRef:
	//               fieldPath: "metadata.name"
	//         - name: "EVENT_LOGS"
	//           value: "recorder,logger"
	//         - name: "baz"
	//           value: "boof"
	//         - name: "foo"
	//           value: "bar"
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{{
		name:    "application/json",
		message: `{"kind":"Received","event":{"data":{"msg":"Hello, 🌎!"},"datacontenttype":"application/json","id":"conformance-0004","source":"//github.com/cloudevents/cloudeventsconformance/yaml/v1.yaml","specversion":"1.0","type":"io.cloudevents.minimum"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["23"],"Content-Type":["application/json; charset=utf-8"],"Host":["recorder-ghpsazde.test-kdmxigkl.svc.cluster.local"],"User-Agent":["Go-http-client/1.1"]},"origin":"10.244.0.8:55854","observer":"recorder-ghpsazde","time":"2021-04-05T22:55:17.447409834Z","sequence":4,"id":""}`,
	}, {
		name:    "application/json; charset=utf-8",
		message: `{"kind":"Received","event":{"data":{"msg":"Hello, 🌎!"},"datacontenttype":"application/json; charset=utf-8","id":"conformance-0004","source":"//github.com/cloudevents/cloudeventsconformance/yaml/v1.yaml","specversion":"1.0","type":"io.cloudevents.minimum"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["23"],"Content-Type":["application/json; charset=utf-8"],"Host":["recorder-ghpsazde.test-kdmxigkl.svc.cluster.local"],"User-Agent":["Go-http-client/1.1"]},"origin":"10.244.0.8:55854","observer":"recorder-ghpsazde","time":"2021-04-05T22:55:17.447409834Z","sequence":4,"id":""}`,
	}, {
		name:    "application/json; charset=utf-8 + string data",
		message: `{"kind":"Received","event":{"data":"Hello!","datacontenttype":"application/json; charset=utf-8","id":"conformance-0004","source":"//github.com/cloudevents/cloudeventsconformance/yaml/v1.yaml","specversion":"1.0","type":"io.cloudevents.minimum"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["23"],"Content-Type":["application/json; charset=utf-8"],"Host":["recorder-ghpsazde.test-kdmxigkl.svc.cluster.local"],"User-Agent":["Go-http-client/1.1"]},"origin":"10.244.0.8:55854","observer":"recorder-ghpsazde","time":"2021-04-05T22:55:17.447409834Z","sequence":4,"id":""}`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventInfo := eventshub.EventInfo{}
			if err := json.Unmarshal([]byte(tt.message), &eventInfo); err != nil {
				t.Errorf("EventInfo that cannot be unmarshalled! \n----\n%s\n----\n%+v\n", tt.message, err)
			}
		})
	}

}

func ExampleForwarder() {
	ctx := testlog.NewContext()
	images := map[string]string{
		"ko://knative.dev/reconciler-test/cmd/eventshub": "uri://a-real-container",
	}
	cfg := map[string]interface{}{
		"name":        "hubhub",
		"serviceName": "hubhub",
		"namespace":   "example",
		"image":       "ko://knative.dev/reconciler-test/cmd/eventshub",
		"sink":        "http://my-svc.cluster.local",
		"envs": map[string]string{
			"foo": "bar",
			"baz": "boof",
		},
	}

	files, err := manifest.ExecuteYAML(ctx, templatesForwarder, images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: serving.knative.dev/v1
	// kind: Service
	// metadata:
	//   name: hubhub
	//   namespace: example
	// spec:
	//   template:
	//     spec:
	//       serviceAccountName: "example"
	//       containers:
	//         - name: eventshub-forwarder
	//           image: uri://a-real-container
	//           imagePullPolicy: "IfNotPresent"
	//           env:
	//             - name: "NAME"
	//               value: hubhub
	//             - name: "NAMESPACE"
	//               value: example
	//             - name: "SINK"
	//               value: http://my-svc.cluster.local
	//             - name: "baz"
	//               value: "boof"
	//             - name: "foo"
	//               value: "bar"
}
