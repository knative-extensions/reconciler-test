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

package manifest_test

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"knative.dev/reconciler-test/pkg/manifest"
	"os"
	"path"
)

func Example_singleExecuteTemplates() {
	images := map[string]string{
		"ko://knative.dev/example/image": "uri://a-real-container",
	}
	cfg := map[string]interface{}{
		"name":      "foo-123",
		"namespace": "example",
	}

	files, err := manifest.ExecuteTemplates(path.Dir("./testdata/single"), "yaml", images, cfg)
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: example.knative.dev/v1
	// kind: Foo
	// metadata:
	//   name: foo-123
	//   namespace: example
	// spec:
	//   foo: bar
	//   image: uri://a-real-container
}

func Example_singleExecuteTestdataYAML() {
	images := map[string]string{
		"ko://knative.dev/example/image": "uri://a-real-container",
	}
	cfg := map[string]interface{}{
		"name":      "foo-123",
		"namespace": "example",
	}

	files, err := manifest.ExecuteYAML(images, cfg, "testdata", "single")
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: example.knative.dev/v1
	// kind: Foo
	// metadata:
	//   name: foo-123
	//   namespace: example
	// spec:
	//   foo: bar
	//   image: uri://a-real-container
}

func Example_multiExecuteTestdataYAML() {
	cfg := map[string]interface{}{
		"name":      "foo-123",
		"namespace": "example",
		"aaaMsg":    "was here",
		"bbbMsg":    "here too",
	}

	files, err := manifest.ExecuteYAML(nil, cfg, "testdata", "multi")
	if err != nil {
		panic(err)
	}

	manifest.OutputYAML(os.Stdout, files)
	// Output:
	// apiVersion: example.knative.dev/v1
	// kind: AAA
	// metadata:
	//   name: aaa-foo-123
	//   namespace: example
	// spec:
	//   aaa: "was here"
	// ---
	// apiVersion: example.knative.dev/v1
	// kind: BBB
	// metadata:
	//   name: bbb-foo-123
	//   namespace: example
	// spec:
	//   bbb: "here too"
}

func Example_singleWithOverrides() {
	images := map[string]string{
		"ko://knative.dev/example/image": "uri://a-real-container",
	}
	cfg := map[string]interface{}{
		"name":      "foo-123",
		"namespace": "example",
		"aaaMsg":    "was here",
	}

	overrides, err := manifest.ExecuteYAML(images, cfg, "testdata", "overrides")
	if err != nil {
		panic(err)
	}

	files, err := manifest.ExecuteYAML(images, cfg, "testdata", "single")
	if err != nil {
		panic(err)
	}

	uls, err := manifest.MergeYAML(files, overrides)
	if err != nil {
		panic(err)
	}

	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	for i, ul := range uls {
		if i > 0 {
			fmt.Println("---")
		}
		err := encoder.Encode(ul.Object)
		if err != nil {
			panic(err)
		}
	}

	// Output:
	// apiVersion: example.knative.dev/v1
	// kind: Foo
	// metadata:
	//   name: foo-123
	//   namespace: example
	// spec:
	//   aaa: was here
	//   foo: bar
	//   image: uri://a-real-container
}
