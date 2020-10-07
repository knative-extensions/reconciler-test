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

package networking

import (
	"path"

	"knative.dev/reconciler-test/pkg/installer"

	"github.com/blang/semver"

	"knative.dev/reconciler-test/pkg/release"

	corev1 "k8s.io/api/core/v1"

	"knative.dev/reconciler-test/pkg/config"
	"knative.dev/reconciler-test/pkg/framework"
	"knative.dev/reconciler-test/pkg/manifest"
)

var (
	Component = &networkingComponent{}

	kourierRelease = release.Release{
		Owner:      "knative-sandbox",
		Repository: "net-kourier",
		Artifacts:  []string{"kourier.yaml"},
	}
)

type networkingComponent struct {
}

var _ framework.Component = (*networkingComponent)(nil)

func (s *networkingComponent) QName() string {
	return "components/networking"
}

func (s *networkingComponent) InstalledVersion(rc framework.ResourceContext) string {
	// TODO: currently there is no way to know which version is installed.

	var obj corev1.Namespace
	_, err := rc.GetOrError("namespace", "kourier", &obj)

	if err != nil {
		return ""
	}

	return "devel"
}

func (s *networkingComponent) Install(rc framework.ResourceContext, gcfg config.Config) {
	cfg, ok := gcfg.(*Config)
	if !ok {
		rc.Errorf("invalid configuration type for %s", s.QName())
	}

	// TODO: check prerequisites (serving)

	if cfg.Version == "devel" {
		rc.Apply(manifest.FromURL(path.Join(cfg.Path, "config")))
		return
	}

	kourierRelease.Install(rc, cfg.Version)

	v, _ := semver.Parse(cfg.Version)
	if hasBuild(v.Build, "kind") {
		// expose kourier service via nodeport
		rc.Apply(manifest.FromString(nodePortService))

		installer.KubectlPatch("configmap", "config-network",
			"-n", "knative-serving",
			"--type", "merge",
			"--patch", `{"data":{"ingress.class":"kourier.ingress.networking.knative.dev"}}`)

		installer.KubectlPatch("configmap", "config-domain",
			"-n", "knative-serving",
			"--type", "merge",
			"--patch", `{"data":{"127.0.0.1.nip.io":""}}`)
	}

}

func hasBuild(build []string, name string) bool {
	if build == nil {
		return false
	}

	for _, b := range build {
		if b == name {
			return true
		}
	}
	return false
}
