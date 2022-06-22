/*
Copyright 2022 The Knative Authors

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

package k8s_test

import (
	"fmt"

	"knative.dev/reconciler-test/pkg/k8s"
	"sigs.k8s.io/yaml"
)

func ExamplePodReference() {
	ref := k8s.PodReference("bar", "foo")
	bytes, err := yaml.Marshal(ref)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
	// Output:
	// apiVersion: v1
	// kind: Pod
	// name: foo
	// namespace: bar
}
