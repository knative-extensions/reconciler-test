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
	"knative.dev/reconciler-test/pkg/config"
)

// Component is set of configuration files that can be installed
// into a cluster
type Component interface {
	// QName returns the fully qualified component name (e.g. components/eventing/sources/github)
	QName() string
}

type ClusterComponent interface {
	Component

	// InstalledVersion returns the installed version of the component.
	// Return empty when the component is not installed
	InstalledVersion(rc ResourceContext) string

	// Install triggers the installation of the component
	Install(rc ResourceContext, gcfg config.Config)
}

type ContainerComponent interface {
	Component

	// Required indicates this component will be used.
	Required(rc ResourceContext, gcfg config.Config)
}
