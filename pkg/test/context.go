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

package test

import (
	"flag"
	"reflect"
	"testing"

	"knative.dev/pkg/test/helpers"
	"knative.dev/reconciler-test/pkg/test/feature"
	"knative.dev/reconciler-test/pkg/test/requirement"
)

// C interface defines the methods required by T to clone and
// setup subtests
type C interface {
	// Copy should return a copy of T
	Copy() C

	// Setup should initialize the context with a test
	//
	// Note: implementations shouldn't need to implement this
	// if they embed a T struct
	Setup(c C, t *testing.T)
}

// T extends testing.T with additional behaviour
// to test various requirement levels & feature states
type T struct {
	*testing.T

	RequirementLevels requirement.Levels
	FeatureStates     feature.States

	// Implemention note:
	//
	// To make authoring tests easier we want to support
	// tests invocations
	//
	// ctx := SomeTestContext{...}
	// ctx.Alpha("name", func(ctx *SomeTestContext) {...})
	//
	// This helps avoid excessive casting in downstream tests
	c C
}

var _ C = (*T)(nil)

// Setup implements the C interface
func (t *T) Setup(c C, test *testing.T) {
	t.c = c
	t.T = test
}

// Copy implements the C interface
func (t *T) Copy() C {
	cpy := *t
	return &cpy
}

// AddFlags adds requirement and feature state flags to the FlagSet.
// The flagset will modify this context instance
//
// Calling AddFlags will also default the requirement level and
// feature states to test everything
func (t *T) AddFlags(fs *flag.FlagSet) {
	if t.RequirementLevels == 0 {
		t.RequirementLevels = requirement.All
	}
	if t.FeatureStates == 0 {
		t.FeatureStates = feature.All
	}

	t.RequirementLevels.AddFlags(fs)
	t.FeatureStates.AddFlags(fs)
}

// Must invokes f as a subtest if the context has the requirement level MUST
func (t *T) Must(name string, f interface{}) bool {
	t.Helper()
	return t.invokeLevel(requirement.Must, name, f)
}

// MustNot invokes f as a subtest only if the context has the requirement level MUST NOT
func (t *T) MustNot(name string, f interface{}) bool {
	t.Helper()
	return t.invokeLevel(requirement.MustNot, name, f)
}

// Should invokes f as a subtest only if the context has the requirement level SHOULD
func (t *T) Should(name string, f interface{}) bool {
	t.Helper()
	return t.invokeLevel(requirement.Should, name, f)
}

// ShouldNot invokes f as a subtest only if the context has the requirement level SHOULD NOT
func (t *T) ShouldNot(name string, f interface{}) bool {
	t.Helper()
	return t.invokeLevel(requirement.ShouldNot, name, f)
}

// May invokes f as a subtest only if the context has the requirement level MAY
func (t *T) May(name string, f interface{}) bool {
	t.Helper()
	return t.invokeLevel(requirement.May, name, f)
}

// Alpha invokes f as a subtest only if the context has the 'Alpha' feature state enabled
func (t *T) Alpha(name string, f interface{}) bool {
	t.Helper()
	return t.invokeFeature(feature.Alpha, name, f)
}

// Beta invokes f as a subtest only if the context has the 'Beta' feature state enabled
func (t *T) Beta(name string, f interface{}) bool {
	t.Helper()
	return t.invokeFeature(feature.Beta, name, f)
}

// Stable invokes f as a subtest only if the context has the 'Stable' feature state enabled
func (t *T) Stable(name string, f interface{}) bool {
	t.Helper()
	return t.invokeFeature(feature.Stable, name, f)
}

// ObjectNameForTest returns a unique resource name based on the test name
func (t *T) ObjectNameForTest() string {
	return helpers.ObjectNameForTest(t.T)
}

// Run invokes f as a subtest
func (t *T) Run(name string, f interface{}) bool {
	t.Helper()
	t.validateCallback(f)

	return t.T.Run(name, func(test *testing.T) {
		t.invoke(f, test)
	})
}

func (t *T) invokeFeature(state feature.States, name string, f interface{}) bool {
	t.Helper()
	t.validateCallback(f)

	return t.T.Run(name, func(test *testing.T) {
		if t.FeatureStates&state == 0 {
			test.Skipf("%s features not enabled for testing", state)
		}
		t.invoke(f, test)
	})
}

func (t *T) invokeLevel(levels requirement.Levels, name string, f interface{}) bool {
	t.Helper()
	t.validateCallback(f)

	return t.T.Run(name, func(test *testing.T) {
		if t.RequirementLevels&levels == 0 {
			test.Skipf("%s requirement not enabled for testing", levels)
		}

		t.invoke(f, test)
	})
}

func (t *T) invoke(f interface{}, test *testing.T) {
	newCtx := t.c.Copy()
	newCtx.Setup(newCtx, test)

	in := []reflect.Value{reflect.ValueOf(newCtx)}
	reflect.ValueOf(f).Call(in)
}

func (t *T) validateCallback(f interface{}) {
	t.Helper()

	if f == nil {
		t.Fatal("callback should not be nil")
	}

	fType := reflect.TypeOf(f)
	if fType.Kind() != reflect.Func {
		t.Fatal("callback should be a function")
	}

	contextType := reflect.TypeOf(t.c)

	if fType.NumIn() != 1 || fType.In(0) != contextType {
		t.Fatalf("callback should take a single argument of %v", contextType)
	}
}
