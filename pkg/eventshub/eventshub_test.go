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
	"os"

	"knative.dev/reconciler-test/pkg/manifest"
)

func Example() {
	images := map[string]string{
		"ko://knative.dev/reconciler-test/cmd/eventshub": "uri://a-real-container",
	}
	cfg := map[string]interface{}{
		"name":      "hubhub",
		"namespace": "example",
		"message":   "Hello, World!",
		"envs": map[string]string{
			"foo": "bar",
			"baz": "boof",
		},
	}

	files, err := manifest.ExecuteLocalYAML(images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: ServiceAccount
	// metadata:
	//   name: hubhub
	//   namespace: example
	// ---
	// apiVersion: rbac.authorization.k8s.io/v1
	// kind: Role
	// metadata:
	//   name: hubhub
	//   namespace: example
	// rules:
	//   - apiGroups: [ "" ]
	//     resources:
	//       - "pods"
	//     verbs:
	//       - "get"
	//       - "list"
	//   - apiGroups: [ "" ]
	//     resources:
	//       - "events"
	//     verbs:
	//       - "*"
	// ---
	// apiVersion: rbac.authorization.k8s.io/v1
	// kind: RoleBinding
	// metadata:
	//   name: hubhub
	//   namespace: example
	// roleRef:
	//   apiGroup: rbac.authorization.k8s.io
	//   kind: Role
	//   name: hubhub
	// subjects:
	//   - kind: ServiceAccount
	//     name: hubhub
	//     namespace: example
	// ---
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
	//   serviceAccountName: "hubhub"
	//   restartPolicy: "Never"
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
