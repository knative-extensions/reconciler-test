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
	//g "github.com/onsi/gomega"
	"fmt"
	"testing"
)

type dummyTest struct {
}

func (d dummyTest) Feature(name string) Test {
	fmt.Print("🎁 " + name + " ")
	return d
}

func (d dummyTest) Alpha() Test {
	fmt.Print("[alpha]")
	return d
}

func (d dummyTest) Beta() Test {
	fmt.Print("[beta]")
	return d
}

func (d dummyTest) Stable() Test {
	fmt.Print("[stable]")
	return d
}

func (d dummyTest) Must(name string) Test {
	fmt.Print("[must]")
	return d
}

func (d dummyTest) Should(name string) Test {
	fmt.Print("[should]")
	return d
}

func (d dummyTest) May(name string) Test {
	fmt.Print("[may]")
	return d
}

func (d dummyTest) Run(fn func(ctx TestContext)) {
	fmt.Println()
	fn(nil)
}

func NewTest(t *testing.T) Test {
	return &dummyTest{}
}

func TestFeature(t *testing.T) {
	NewTest(t).Feature("some feature").Run(SomeFeature)
}

func SomeFeature(tc TestContext) {
	//tc.Helper()
	//tc.Expect(true).To(g.BeTrue())
}

func TestAlphaFeature(t *testing.T) {
	NewTest(t).Feature("some feature").Alpha().Run(SomeFeature)
}
