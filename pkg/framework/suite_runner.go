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
	_ "knative.dev/pkg/system/testing"
)

var (
	config *BaseConfig
	cfg    *rest.Config
)

type suite struct {
	m *testing.M
}

func newSuite(m *testing.M) Suite {
	return &suite{m: m}
}

func (s *suite) Configure(def Config) Suite {
	def.SetDefaults()

	// TODO: read config file

	err := gflag.ParseToDef(def)
	if err != nil {
		panic(err)
	}

	flag.Parse()

	config = def.GetBaseConfig()
	return s
}

func (s *suite) Require(component Component) Suite {
	// TODO: delegate to the component. Must first define configuration
	return s
}

func (s *suite) Run() {
	if config == nil {
		// Use default configuration
		s.Configure(&BaseConfig{})
	}

	cfg = s.enableInjection()

	os.Exit(s.m.Run())
}

func (s *suite) enableInjection() *rest.Config {
	ctx := signals.NewContext()

	cfg, err := sharedmain.GetConfig(config.serverURL, config.KubeConfig)
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
