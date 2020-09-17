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

type ComponentScope string

const (
	ComponentScopeResource  = ComponentScope("resource")
	ComponentScopeNamespace = ComponentScope("namespace")
	ComponentScopeCluster   = ComponentScope("cluster")
)

// Component is set of configuration files that can be installed
// into a cluster
type Component interface {
	// Scope returns the component scope
	Scope() ComponentScope

	// Build builds and publishes the container image(s) required to
	// install the component.
	// Do nothing if container image(s) are pre-built.
	BuildOrFail()

	// Install deploys the component onto the cluster identified by rc.
	// Do nothing if the component of the same version is already installed.
	InstallOrFail(rc ResourceContext, config Config)
}
