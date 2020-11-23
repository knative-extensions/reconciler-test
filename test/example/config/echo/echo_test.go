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

package echo_test

import (
	"os"

	"knative.dev/reconciler-test/pkg/manifest"
)

func Example() {
	images := map[string]string{
		"ko://knative.dev/reconciler-test/test/example/cmd/echo": "uri://a-real-container",
	}
	cfg := map[string]interface{}{
		"namespace": "example",
		"message":   "Hello, World!",
	}

	files, err := manifest.ExecuteLocalYAML(images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: batch/v1
	// kind: Job
	// metadata:
	//   name: echo
	//   namespace: example
	// spec:
	//   backoffLimit: 0
	//   parallelism: 1
	//   template:
	//     spec:
	//       restartPolicy: Never
	//       containers:
	//         - name: echo
	//           image: uri://a-real-container
	//           env:
	//             - name: ECHO
	//               value: "Hello, World!"
}
