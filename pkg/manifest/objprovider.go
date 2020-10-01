/*
 * Copyright 2020 The Knative Authors
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

package manifest

import (
	"io/ioutil"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

type objProvider struct {
	obj runtime.Object
}

func (p *objProvider) GetPath() (string, error) {
	f, err := ioutil.TempFile("", "manifest-*.yaml")
	if err != nil {
		return "", err
	}
	defer f.Close()

	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{Yaml: true})
	err = serializer.Encode(p.obj, f)
	if err != nil {
		return "", nil
	}
	return f.Name(), nil
}

func (p *objProvider) Recursive() bool {
	return false
}
