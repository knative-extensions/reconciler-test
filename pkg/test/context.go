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

// Context interface defines the methods required by a test context
type Context interface {
	// Copy should return a copy of the existing context
	Copy() Context

	// Setup should initialize the context with a test
	//
	// Note: implementations shouldn't need to implement this
	// if they embed a BaseContext struct
	Setup(c Context, t *testing.T)
}

// BaseContext extends testing.T with additional behaviour
// to test various requirement levels & feature states
type BaseContext struct {
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
	c Context
}

var _ Context = (*BaseContext)(nil)

// Setup implements the Context interface
func (b *BaseContext) Setup(c Context, t *testing.T) {
	b.c = c
	b.T = t
}

// Copy implements the Context interface
func (b *BaseContext) Copy() Context {
	cpy := *b
	return &cpy
}

// AddFlags adds requirement and feature state flags to the FlagSet.
// The flagset will modify this context instance
//
// Calling AddFlags will also default the requirement level and
// feature states to test everything
func (b *BaseContext) AddFlags(fs *flag.FlagSet) {
	if b.RequirementLevels == 0 {
		b.RequirementLevels = requirement.All
	}
	if b.FeatureStates == 0 {
		b.FeatureStates = feature.All
	}

	b.RequirementLevels.AddFlags(fs)
	b.FeatureStates.AddFlags(fs)
}

// Must invokes f as a subtest if the context has the requirement level MUST
func (b *BaseContext) Must(name string, f interface{}) bool {
	b.Helper()
	return b.invokeLevel(requirement.Must, name, f)
}

// MustNot invokes f as a subtest only if the context has the requirement level MUST NOT
func (b *BaseContext) MustNot(name string, f interface{}) bool {
	b.Helper()
	return b.invokeLevel(requirement.MustNot, name, f)
}

// Should invokes f as a subtest only if the context has the requirement level SHOULD
func (b *BaseContext) Should(name string, f interface{}) bool {
	b.Helper()
	return b.invokeLevel(requirement.Should, name, f)
}

// ShouldNot invokes f as a subtest only if the context has the requirement level SHOULD NOT
func (b *BaseContext) ShouldNot(name string, f interface{}) bool {
	b.Helper()
	return b.invokeLevel(requirement.ShouldNot, name, f)
}

// May invokes f as a subtest only if the context has the requirement level MAY
func (b *BaseContext) May(name string, f interface{}) bool {
	b.Helper()
	return b.invokeLevel(requirement.May, name, f)
}

// Alpha invokes f as a subtest only if the context has the 'Alpha' feature state enabled
func (b *BaseContext) Alpha(name string, f interface{}) bool {
	b.Helper()
	return b.invokeFeature(feature.Alpha, name, f)
}

// Beta invokes f as a subtest only if the context has the 'Beta' feature state enabled
func (b *BaseContext) Beta(name string, f interface{}) bool {
	b.Helper()
	return b.invokeFeature(feature.Beta, name, f)
}

// Stable invokes f as a subtest only if the context has the 'Stable' feature state enabled
func (b *BaseContext) Stable(name string, f interface{}) bool {
	b.Helper()
	return b.invokeFeature(feature.Stable, name, f)
}

// ObjectNameForTest returns a unique resource name based on the test name
func (b *BaseContext) ObjectNameForTest() string {
	return helpers.ObjectNameForTest(b.T)
}

// Run invokes f as a subtest
func (b *BaseContext) Run(name string, f interface{}) bool {
	b.Helper()
	b.validateCallback(f)

	return b.T.Run(name, func(t *testing.T) {
		b.invoke(f, t)
	})
}

func (b *BaseContext) invokeFeature(state feature.States, name string, f interface{}) bool {
	b.Helper()
	b.validateCallback(f)

	return b.T.Run(name, func(t *testing.T) {
		if b.FeatureStates&state == 0 {
			t.Skipf("%s features not enabled for testing", state)
		}
		b.invoke(f, t)
	})
}

func (b *BaseContext) invokeLevel(levels requirement.Levels, name string, f interface{}) bool {
	b.Helper()
	b.validateCallback(f)

	return b.T.Run(name, func(t *testing.T) {
		if b.RequirementLevels&levels == 0 {
			t.Skipf("%s requirement not enabled for testing", levels)
		}

		b.invoke(f, t)
	})
}

func (b *BaseContext) invoke(f interface{}, t *testing.T) {
	newCtx := b.c.Copy()
	newCtx.Setup(newCtx, t)

	in := []reflect.Value{reflect.ValueOf(newCtx)}
	reflect.ValueOf(f).Call(in)
}

func (b *BaseContext) validateCallback(f interface{}) {
	b.Helper()

	if f == nil {
		b.Fatal("callback should not be nil")
	}

	fType := reflect.TypeOf(f)
	if fType.Kind() != reflect.Func {
		b.Fatal("callback should be a function")
	}

	contextType := reflect.TypeOf(b.c)

	if fType.NumIn() != 1 || fType.In(0) != contextType {
		b.Fatalf("callback should take a single argument of %v", contextType)
	}
}
