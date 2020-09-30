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

package serving

import (
	"path"

	"knative.dev/reconciler-test/pkg/release"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"knative.dev/reconciler-test/pkg/config"
	"knative.dev/reconciler-test/pkg/framework"
	"knative.dev/reconciler-test/pkg/manifest"
)

var (
	Component = &servingComponent{}

	servingRelease = release.Release{
		Owner:      "knative",
		Repository: "serving",
		Artifacts:  []string{"serving-crds.yaml", "serving-core.yaml"},
	}
)

type servingComponent struct {
}

var _ framework.Component = (*servingComponent)(nil)

func (s *servingComponent) QName() string {
	return "components/serving"
}

func (s *servingComponent) InstalledVersion(rc framework.ResourceContext) string {
	rc = framework.NewResourceContext(rc, "knative-serving")
	var obj apiextensionsv1.CustomResourceDefinition
	_, err := rc.GetOrError("customresourcedefinitions", "services.serving.knative.dev", &obj)

	if err != nil {
		return ""
	}
	if v, ok := obj.Labels["serving.knative.dev/release"]; ok {
		return v
	}
	return ""
}

func (s *servingComponent) Install(rc framework.ResourceContext, gcfg config.Config) {
	cfg, ok := gcfg.(*Config)
	if !ok {
		rc.Errorf("invalid configuration type for %s", s.QName())
	}

	if cfg.Version == "devel" {
		rc.Apply(manifest.FromURL(path.Join(cfg.Path, "config")))
		return
	}

	servingRelease.Install(rc, cfg.Version)
}
