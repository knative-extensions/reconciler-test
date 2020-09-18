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

import "testing"

// Test defines functions for configuring and running a single test case
type Test interface {

	// Feature labels the test case as exercising a feature
	Feature(name string) Test

	// --- Stability markers

	// Alpha marks the feature as alpha
	Alpha() Test

	// Beta marks the feature as beta
	Beta() Test

	// Stable marks the feature as stable
	Stable() Test

	// --- Requirements markers

	// Must marks the feature as must have
	Must(name string) Test

	// Must marks the feature as should have
	Should(name string) Test

	// Must marks the feature as may have
	May(name string) Test

	// --- Runners

	// Run the test within the given context
	Run(fn func(tc TestContext))
}

// NewTest creates a new test case
func NewTest(t *testing.T) Test {
	return newTest(t)
}
