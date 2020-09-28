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

import "knative.dev/reconciler-test/pkg/config"

type ComponentScope string

const (
	ComponentScopeResource  = ComponentScope("resource")
	ComponentScopeNamespace = ComponentScope("namespace")
	ComponentScopeCluster   = ComponentScope("cluster")
)

// Component is set of configuration files that can be installed
// into a cluster
type Component interface {
	// Scope returns the component scope.
	Scope() ComponentScope

	// Required marks the component as being required.
	Required(rc ResourceContext, cfg config.Config)
}
