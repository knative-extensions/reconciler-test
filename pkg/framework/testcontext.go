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
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/onsi/gomega"
)

// TestContext is the context when running a test case
type TestContext interface {
	// Create a resource from the given object (or fail)
	CreateOrFail(obj runtime.Object)

	// CreateFromYAMLOrFail creates a resource from the given YAML specification (or fail)
	CreateFromYAMLOrFail(yaml string)

	// --- testing.T

	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Skip(args ...interface{})
	SkipNow()
	Skipf(format string, args ...interface{})
	Skipped() bool

	// --- Assertion

	// Gomega assertion
	gomega.Gomega
}
