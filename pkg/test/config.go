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

	"k8s.io/client-go/kubernetes"
	"knative.dev/reconciler-test/pkg/test/environment"
	"knative.dev/reconciler-test/pkg/test/feature"
	"knative.dev/reconciler-test/pkg/test/requirement"
)

type Config interface {
	Features() feature.States
	Environment() environment.Settings
	Requirements() requirement.Levels
	KubeClient() kubernetes.Interface
	ResetClients()
}

type BaseConfig struct {
	requirement.Levels
	environment.Settings
	feature.States

	kube kubernetes.Interface
}

func (c *BaseConfig) AddFlags(fs *flag.FlagSet) {
	c.Levels.AddFlags(fs)
	c.Settings.AddFlags(fs)
	c.States.AddFlags(fs)
}

func (c *BaseConfig) Features() feature.States {
	return c.States
}

func (c *BaseConfig) Environment() environment.Settings {
	return c.Settings
}

func (c *BaseConfig) Requirements() requirement.Levels {
	return c.Levels
}

func (c *BaseConfig) ResetClients() {
	c.kube = kubernetes.NewForConfigOrDie(c.Environment().ClientConfig())
}

func (c *BaseConfig) KubeClient() kubernetes.Interface {
	return c.kube
}
