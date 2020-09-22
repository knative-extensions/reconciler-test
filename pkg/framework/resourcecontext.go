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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"knative.dev/reconciler-test/pkg/manifest"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"knative.dev/pkg/injection/clients/dynamicclient"

	"k8s.io/apimachinery/pkg/runtime"
)

// ResourceContext is the context in which resources are managed.
type ResourceContext interface {
	context.Context

	// Namespace returns the current namespace
	Namespace() string

	// ImageName returns the image name corresponding to the given Go package name
	ImageName(packageName string) string

	// Create a resource from the given object (or fail)
	CreateOrFail(obj runtime.Object)

	// CreateFromYAMLOrFail creates resources from the given YAML specification (or fail)
	CreateFromYAMLOrFail(yaml string)

	// CreateFromURIOrFail creates resources from the given URi (or fail)
	// 1. pathname = path to a file --> parses that file.
	// 2. pathname = path to a directory, recursive = false --> parses all files in
	//    that directory.
	// 3. pathname = path to a directory, recursive = true --> parses all files in
	//    that directory and it's descendants
	// 4. pathname = url --> fetches the contents of that URL and parses them as YAML.
	// 5. pathname = combination of all previous cases, the string can contain
	//    multiple records (file, directory or url) separated by comma
	CreateFromURIOrFail(uri string, recursive bool)

	// Delete deletes the resource specified in the given YAML
	DeleteFromYAML(yaml string) error

	// Delete deletes the resource specified in the given YAML (or fail)
	DeleteFromYAMLOrFail(yaml string)

	// TODO: Get, Update, Apply

	// --- Failures. Subset of testing.T

	Helper()
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

// --- Default implementation

type resourceContextImpl struct {
	context   context.Context // I know
	namespace string
}

func (c *resourceContextImpl) Namespace() string {
	return c.namespace
}

func (c *resourceContextImpl) ImageName(packageName string) string {
	repository := baseConfig.ImageRepository
	if repository == "ko" {
		repository = os.Getenv("KO_DOCKER_REPO")
		if repository == "" {
			panic("error: KO_DOCKER_REPO environment variable is unset")
		}
	}
	parts := strings.Split(packageName, "/")
	return fmt.Sprintf("%s/%s", repository, parts[len(parts)-1])
}

func (c *resourceContextImpl) CreateOrFail(obj runtime.Object) {
	c.Helper()
	gvr, _ := meta.UnsafeGuessKindToResource(obj.GetObjectKind().GroupVersionKind())
	u, err := toUnstructured(obj)
	if err != nil {
		c.Fatal(err)
	}
	// TODO: retrywebhook as an option
	if clusterScoped(gvr) {
		if _, err = dynamicclient.Get(c).Resource(gvr).Create(c, u, metav1.CreateOptions{}); err != nil {

			c.Fatal(err)
		}
	} else {
		ns := u.GetNamespace()
		if ns == "" {
			ns = c.namespace
			if ns == "" {
				fmt.Println(gvr)
				c.Fatalf("unbound namespace for resource %s/%s", u.GroupVersionKind().String(), u.GetName())
			}
		}

		if _, err = dynamicclient.Get(c).Resource(gvr).Namespace(ns).Create(c, u, metav1.CreateOptions{}); err != nil {
			c.Fatal(err)
		}
	}
}

func (c *resourceContextImpl) CreateFromYAMLOrFail(yamlSpec string) {
	c.Helper()

	us, err := manifest.ParseString(yamlSpec)
	if err != nil {
		c.Fatal(err)
	}

	for _, u := range us {
		c.CreateOrFail(&u)
	}
}

func (c *resourceContextImpl) CreateFromURIOrFail(pathname string, recursive bool) {
	c.Helper()

	us, err := manifest.Parse(pathname, recursive)
	if err != nil {
		c.Fatal(err)
	}

	for _, u := range us {
		c.CreateOrFail(&u)
	}
}

func (c *resourceContextImpl) DeleteFromYAML(yamlSpec string) error {
	c.Helper()
	decoder := yaml.NewYAMLToJSONDecoder(strings.NewReader(yamlSpec))

	out := unstructured.Unstructured{}
	if err := decoder.Decode(&out); err != nil {
		c.Fatal(err)
	}

	gvr, _ := meta.UnsafeGuessKindToResource(out.GroupVersionKind())
	if clusterScoped(gvr) {
		return dynamicclient.Get(c).Resource(gvr).Delete(c, out.GetName(), metav1.DeleteOptions{})
	} else {
		return dynamicclient.Get(c).Resource(gvr).Namespace(c.namespace).Delete(c, out.GetName(), metav1.DeleteOptions{})
	}
	return nil

}

func (c *resourceContextImpl) DeleteFromYAMLOrFail(yamlSpec string) {
	c.Helper()
	if err := c.DeleteFromYAML(yamlSpec); err != nil {
		c.Fatal(err)
	}
}

// --- Failures

func (c *resourceContextImpl) Helper() {
}

func (c *resourceContextImpl) Error(args ...interface{}) {
	panic(fmt.Sprintln(args...))
}

func (c *resourceContextImpl) Errorf(format string, args ...interface{}) {
	c.Error(fmt.Sprintf(format, args...))
}

func (c *resourceContextImpl) Fatal(args ...interface{}) {
	c.Error(args...)
}

func (c *resourceContextImpl) Fatalf(format string, args ...interface{}) {
	c.Errorf(format, args...)
}

// --- context.Context

func (c *resourceContextImpl) Deadline() (deadline time.Time, ok bool) {
	return c.context.Deadline()
}

func (c *resourceContextImpl) Done() <-chan struct{} {
	return c.context.Done()
}

func (c *resourceContextImpl) Err() error {
	return c.context.Err()
}

func (c *resourceContextImpl) Value(key interface{}) interface{} {
	return c.context.Value(key)
}

// --- utils

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
	"":                             {"namespaces": true},
	"rbac.authorization.k8s.io":    {"clusterroles": true, "clusterrolebindings": true},
	"apiextensions.k8s.io":         {"customresourcedefinitions": true},
	"admissionregistration.k8s.io": {"mutatingwebhookconfigurations": true, "validatingwebhookconfigurations": true},

	// TODO: add more
}

func clusterScoped(gvr schema.GroupVersionResource) bool {
	if r, ok := gr[gvr.Group]; ok {
		_, ok := r[gvr.Resource]
		return ok
	}
	return false
}
