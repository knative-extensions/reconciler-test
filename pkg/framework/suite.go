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

	"knative.dev/reconciler-test/pkg/config"
)

// Suite represents a collection of test cases.
// Must be instantiated in TestMain.
type Suite interface {
	// Configure assembles the suite configuration by
	// - reading the configuration file in the test directory, up to the root project
	// - overriding config values with the one provided on the command line
	//
	// Not calling this function is equivalent to calling it with BaseConfig.
	Configure(def config.Config) Suite

	// Require indicates the given component is needed to
	// run test cases.
	Require(component Component) Suite

	// Run runs the tests
	Run()
}

// NewSuite creates a new test suite
func NewSuite(m *testing.M) Suite {
	return newSuite(m)
}
