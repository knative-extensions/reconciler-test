/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package github

import (
	"fmt"

	"knative.dev/reconciler-test/pkg/manifest"

	"knative.dev/reconciler-test/pkg/config"
	"knative.dev/reconciler-test/pkg/framework"
)

const (
	artifactURLTemplate = "https://github.com/knative/eventing-contrib/releases/download/v%s/github.yaml"
)

var (
	Component = &githubComponent{}
)

type githubComponent struct {
}

var _ framework.Component = (*githubComponent)(nil)

// Scope returns the component scope
func (s *githubComponent) Scope() framework.ComponentScope {
	return framework.ComponentScopeCluster
}

func (s *githubComponent) Required(rc framework.ResourceContext, cfg config.Config) {
	ghcfg := config.GetConfig(cfg, "components/eventing/sources/github").(GithubConfig)

	// TODO: validate configuration
	// TODO: check cluster for existing source
	artifactURL := fmt.Sprintf(artifactURLTemplate, ghcfg.Version)
	rc.Logf("installing GitHubSource release ", ghcfg.Version)
	rc.Apply(manifest.FromURL(artifactURL))
}
