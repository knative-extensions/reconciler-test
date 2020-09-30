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

type ComponentManager struct {
	components map[string]Component
}

func NewComponentManager() ComponentManager {
	return ComponentManager{
		components: make(map[string]Component),
	}
}

// Required registers component is required.
func (cm *ComponentManager) Required(rc ResourceContext, component Component, cfg config.Config) {
	if _, ok := cm.components[component.QName()]; ok {
		return
	}

	cm.components[component.QName()] = component
	if cs, ok := component.(ClusterComponent); ok {
		csrequired(rc, cs, cfg)
	}

	if cc, ok := component.(ContainerComponent); ok {
		ccrequired(rc, cc, cfg)
	}
}

func (cm *ComponentManager) Wait(rc ResourceContext) {
	// TODO: wait for all components to be ready
}

func ccrequired(rc ResourceContext, component ContainerComponent, cfg config.Config) {
	ccfg := config.GetConfig(cfg, component.QName())
	if ccfg == nil {
		rc.Errorf("☠️ missing %s configuration", component.QName())
	}

	if errs := ccfg.Validate(); errs != nil {
		rc.Errorf("☠️ invalid configuration: %v", errs.ViaField(component.QName()))
	}

	component.Required(rc, ccfg.(config.Config))
}

func csrequired(rc ResourceContext, component ClusterComponent, cfg config.Config) {
	ccfg := config.GetConfig(cfg, component.QName())
	if ccfg == nil {
		rc.Errorf("☠️ missing %s configuration", component.QName())
	}

	if errs := ccfg.Validate(); errs != nil {
		rc.Errorf("☠️ invalid configuration %v", errs.ViaField(component.QName()))
	}

	vspec, ok := config.GetConfig(ccfg.(config.Config), "VersionSpec").(*config.VersionSpec)
	if !ok {
		rc.Errorf("☠️ component %s configuration does not embed config.VersionSpec", component.QName())
	}

	iv := component.InstalledVersion(rc)
	diff := vspec.Compare(iv)

	switch diff {
	case config.CompareInRange:
		rc.Logf("✅ component %s@%v already installed (required %v)", component.QName(), vspec.ActualVersion, vspec.Require)
	case config.CompareDevel:
		rc.Logf("🏃 installing component %s from source code", component.QName())
		component.Install(rc, ccfg.(config.Config))
	case config.CompareOutOfRange:
		rc.Errorf("☠️ installed component %s@%s does not match required version @%s", component.QName(), vspec.ActualVersion, vspec.Version)
	case config.CompareInvalidVersion:
		rc.Errorf("️️☠️ invalid version %s", iv)
	case config.CompareEmptyVersion:
		rc.Logf("🏃 installing component %s@%s", component.QName(), vspec.Version)
		component.Install(rc, ccfg.(config.Config))
	}
}
