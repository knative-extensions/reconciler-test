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

package eventshub_test

import (
	"embed"
	"os"

	"knative.dev/reconciler-test/pkg/manifest"
)

//go:embed *.yaml
var templates embed.FS

func Example() {
	files, err := manifest.ExecuteYAML(templates, nil,
		map[string]interface{}{
			"namespace": "example",
		})
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: v1
	// kind: ServiceAccount
	// metadata:
	//   name: example
	//   namespace: example
	// ---
	// apiVersion: rbac.authorization.k8s.io/v1
	// kind: Role
	// metadata:
	//   name: example
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
	//   name: example
	//   namespace: example
	// roleRef:
	//   apiGroup: rbac.authorization.k8s.io
	//   kind: Role
	//   name: example
	// subjects:
	//   - kind: ServiceAccount
	//     name: example
	//     namespace: example
}
