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

package framework

import (
	"encoding/json"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/pkg/injection/clients/dynamicclient"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (c *testContextImpl) Namespace() string {
	return c.namespace
}

func (c *testContextImpl) CreateOrFail(obj runtime.Object) {
	c.t.Helper()
	gvr, _ := meta.UnsafeGuessKindToResource(obj.GetObjectKind().GroupVersionKind())
	u, err := toUnstructured(obj)
	if err != nil {
		c.t.Fatal(err)
	}
	// TODO: retrywebhook as an option
	if clusterScoped(gvr) {
		if _, err = dynamicclient.Get(c).Resource(gvr).Create(c, u, metav1.CreateOptions{}); err != nil {
			c.t.Fatal(err)
		}
	} else {
		if _, err = dynamicclient.Get(c).Resource(gvr).Namespace(c.namespace).Create(c, u, metav1.CreateOptions{}); err != nil {
			c.t.Fatal(err)
		}
	}
}

func (c *testContextImpl) CreateFromYAMLOrFail(yamlSpec string) {
	c.t.Helper()
	decoder := yaml.NewYAMLToJSONDecoder(strings.NewReader(yamlSpec))

	out := unstructured.Unstructured{}
	if err := decoder.Decode(&out); err != nil {
		c.t.Fatal(err)
	}

	gvr, _ := meta.UnsafeGuessKindToResource(out.GroupVersionKind())
	if clusterScoped(gvr) {
		if _, err := dynamicclient.Get(c).Resource(gvr).Create(c, &out, metav1.CreateOptions{}); err != nil {
			c.t.Fatal(err)
		}
	} else {
		if _, err := dynamicclient.Get(c).Resource(gvr).Namespace(c.namespace).Create(c, &out, metav1.CreateOptions{}); err != nil {
			c.t.Fatal(err)
		}
	}
}

func (c *testContextImpl) DeleteFromYAML(yamlSpec string) error {
	c.t.Helper()
	decoder := yaml.NewYAMLToJSONDecoder(strings.NewReader(yamlSpec))

	out := unstructured.Unstructured{}
	if err := decoder.Decode(&out); err != nil {
		c.t.Fatal(err)
	}

	gvr, _ := meta.UnsafeGuessKindToResource(out.GroupVersionKind())
	if clusterScoped(gvr) {
		return dynamicclient.Get(c).Resource(gvr).Delete(c, out.GetName(), metav1.DeleteOptions{})
	} else {
		return dynamicclient.Get(c).Resource(gvr).Namespace(c.namespace).Delete(c, out.GetName(), metav1.DeleteOptions{})
	}
	return nil

}

func (c *testContextImpl) DeleteFromYAMLOrFail(yamlSpec string) {
	c.t.Helper()
	if err := c.DeleteFromYAML(yamlSpec); err != nil {
		c.t.Fatal(err)
	}
}

func (c *testContextImpl) Deadline() (deadline time.Time, ok bool) {
	return c.context.Deadline()
}

func (c *testContextImpl) Done() <-chan struct{} {
	return c.context.Done()
}

func (c *testContextImpl) Err() error {
	return c.context.Err()
}

func (c *testContextImpl) Value(key interface{}) interface{} {
	return c.context.Value(key)
}

func toUnstructured(desired runtime.Object) (*unstructured.Unstructured, error) {
	// Convert desired to unstructured.Unstructured
	b, err := json.Marshal(desired)
	if err != nil {
		return nil, err
	}
	ud := &unstructured.Unstructured{}
	if err := json.Unmarshal(b, ud); err != nil {
		return nil, err
	}
	return ud, nil
}

var gr = map[string]map[string]bool{
	"": {"namespaces": true}, // TODO: add more
}

func clusterScoped(gvr schema.GroupVersionResource) bool {
	if r, ok := gr[gvr.Group]; ok {
		_, ok := r[gvr.Resource]
		return ok
	}
	return false
}
