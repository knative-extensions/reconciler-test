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
	"fmt"
	"testing"
)

type test struct {
	t *testing.T
}

func newTest(t *testing.T) Test {
	return &test{t: t}
}

func (t *test) Feature(name string) Test {
	fmt.Print("🎁 " + name + " ")
	return t
}

func (t *test) Alpha() Test {
	fmt.Print("[alpha]")
	return t
}

func (t *test) Beta() Test {
	fmt.Print("[beta]")
	return t
}

func (t *test) Stable() Test {
	fmt.Print("[stable]")
	return t
}

func (t *test) Must(name string) Test {
	fmt.Print("[must]")
	return t
}

func (t *test) Should(name string) Test {
	fmt.Print("[should]")
	return t
}

func (t *test) May(name string) Test {
	fmt.Print("[may]")
	return t
}

func (t *test) Run(fn func(ctx TestContext)) {
	fmt.Println()
	fn(nil)
}
