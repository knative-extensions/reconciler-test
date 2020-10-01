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
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/octago/sflags/gen/gflag"
	"k8s.io/client-go/rest"

	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/system"

	"knative.dev/reconciler-test/pkg/config"
	"knative.dev/reconciler-test/pkg/installer"
)

func init() {
	// Do not import pkg/test
	if ns := os.Getenv(system.NamespaceEnvKey); ns != "" {
		return
	}
	os.Setenv(system.NamespaceEnvKey, "knative-testing")
}

var (
	fullConfig config.Config
	baseConfig *BaseConfig
	cfg        *rest.Config
	rc         *resourceContextImpl
)

type suite struct {
	m *testing.M
}

func newSuite(m *testing.M) Suite {
	return &suite{m: m}
}

func (s *suite) Configure(def config.Config) Suite {
	config.ParseConfigFile(def)

	// Overrides loaded configuration
	err := gflag.ParseToDef(def)
	if err != nil {
		panic(err)
	}

	flag.Parse()

	fullConfig = def

	bcfg, ok := config.GetConfig(def, "BaseConfig").(BaseConfig)
	if !ok {
		panic("Configuration must embed framework.BaseConfig")
	}
	baseConfig = &bcfg

	cfg = s.enableInjection()

	rc = &resourceContextImpl{
		context:   withInjection(context.Background()),
		namespace: "",
	}

	return s
}

func (s *suite) Require(component Component) Suite {
	s.mayConfigure()

	component.Required(rc, fullConfig)

	return s
}

func (s *suite) Run() {
	s.mayConfigure()

	s.prepareComponents()

	os.Exit(s.m.Run())
}

func (s *suite) mayConfigure() {
	if baseConfig == nil {
		s.Configure(&BaseConfig{})
	}
}

func (s *suite) prepareComponents() {
	if baseConfig.BuildImages {
		fmt.Println("building and publishing images")
		installer.ProduceImages()
	}
}

func (s *suite) enableInjection() *rest.Config {
	ctx := signals.NewContext()

	cfg, err := sharedmain.GetConfig(baseConfig.ServerURL, baseConfig.KubeConfig)
	if err != nil {
		panic(err)
	}

	ctx = injection.WithConfig(ctx, cfg)
	ctx, informers := injection.Default.SetupInformers(ctx, cfg)

	// Start the injection clients and informers.
	go func(ctx context.Context) {
		if err := controller.StartInformers(ctx.Done(), informers...); err != nil {
			panic(fmt.Sprintf("Failed to start informers - %s", err))
		}
		<-ctx.Done()
	}(ctx)

	return cfg
}

func withInjection(ctx context.Context) context.Context {
	ctx = injection.WithConfig(ctx, cfg)
	ctx, _ = injection.Default.SetupInformers(ctx, cfg)
	// do not start informers.
	return ctx

}
