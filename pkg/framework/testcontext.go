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
	"testing"

	"github.com/onsi/gomega"
)

// TestContext is the context when running a test case
type TestContext interface {
	ResourceContext

	// --- The rest of testing.T

	Fail()
	FailNow()
	Failed() bool
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

// --- Default implementation

type testContextImpl struct {
	resourceContextImpl

	t *testing.T
	*gomega.WithT
}

// --- testing.T wrapper

func (c *testContextImpl) Error(args ...interface{}) {
	c.t.Error(args...)
}

func (c *testContextImpl) Errorf(format string, args ...interface{}) {
	c.t.Errorf(format, args...)
}

func (c *testContextImpl) Fail() {
	c.t.Fail()
}

func (c *testContextImpl) FailNow() {
	c.t.FailNow()
}

func (c *testContextImpl) Failed() bool {
	return c.t.Failed()
}

func (c *testContextImpl) Fatal(args ...interface{}) {
	c.t.Fatal(args...)
}

func (c *testContextImpl) Fatalf(format string, args ...interface{}) {
	c.t.Fatalf(format, args...)
}

func (c *testContextImpl) Helper() {
	c.t.Helper()
}

func (c *testContextImpl) Log(args ...interface{}) {
	c.t.Log(args...)
}

func (c *testContextImpl) Logf(format string, args ...interface{}) {
	c.t.Logf(format, args...)
}

func (c *testContextImpl) Name() string {
	return c.t.Name()
}

func (c *testContextImpl) Skip(args ...interface{}) {
	c.t.Skip(args...)
}

func (c *testContextImpl) SkipNow() {
	c.t.SkipNow()
}

func (c *testContextImpl) Skipf(format string, args ...interface{}) {
	c.t.Skipf(format, args...)
}

func (c *testContextImpl) Skipped() bool {
	return c.t.Skipped()
}
