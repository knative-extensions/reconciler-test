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

package environment

import (
	"bytes"
	"flag"
	"os"
	"os/user"
	"path"
	"text/template"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Settings struct {
	Cluster    string // K8s cluster (defaults to cluster in kubeconfig)
	KubeConfig string // Path to kubeconfig (defaults to ./kube/config)
	Namespace  string // K8s namespace (blank by default, to be overwritten by test suite)

	IngressEndpoint  string // Host to use for ingress endpoint
	IngressName      string // K8s Service name acting as an ingress
	IngressNamespace string // K8s Service namespace acting as an ingress

	ImageTemplate string // Template to build the image reference (defaults to {{.Repository}}/{{.Name}}:{{.Tag}})
	ImageRepo     string // Image repository (defaults to $KO_DOCKER_REPO)
	ImageTag      string // Tag for test images

	SpoofRequestInterval time.Duration // SpoofRequestInterval is the interval between requests in SpoofingClient
	SpoofRequestTimeout  time.Duration // SpoofRequestTimeout is the timeout for polling requests in SpoofingClient
}

func (s *Settings) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&s.Cluster, "env.cluster", "",
		"Provide the cluster to test against. Defaults to the current cluster in kubeconfig.")

	// Use KUBECONFIG if available
	defaultKubeconfig := os.Getenv("KUBECONFIG")

	// If KUBECONFIG env var isn't set then look for $HOME/.kube/config
	if defaultKubeconfig == "" {
		if usr, err := user.Current(); err == nil {
			defaultKubeconfig = path.Join(usr.HomeDir, ".kube/config")
		}
	}

	// Allow for --kubeconfig on the cmd line to override the above logic
	fs.StringVar(&s.KubeConfig, "env.kubeconfig", defaultKubeconfig,
		"Provide the path to the `kubeconfig` file you'd like to use for these tests. The `current-context` will be used.")

	fs.StringVar(&s.Namespace, "env.namespace", "",
		"Provide the namespace you would like to use for these tests.")

	fs.StringVar(&s.IngressEndpoint, "env.ingress.endpoint", "",
		"Provide a static endpoint url to the ingress server used during tests.")

	fs.StringVar(&s.IngressName, "env.ingress.name", "",
		"Provide a k8s service name as a target ingress used during tests.")

	fs.StringVar(&s.IngressName, "env.ingress.namespace", "",
		"Provide a k8s service namespace as a target ingress used during tests.")

	fs.DurationVar(&s.SpoofRequestInterval, "env.spoof.interval", 1*time.Second,
		"Provide an interval between requests for the SpoofingClient")

	fs.DurationVar(&s.SpoofRequestTimeout, "env.spoof.timeout", 5*time.Minute,
		"Provide a request timeout for the SpoofingClient")

	defaultRepo := os.Getenv("KO_DOCKER_REPO")
	fs.StringVar(&s.ImageRepo, "env.image.repo", defaultRepo,
		"Provide the uri of the docker repo you have uploaded the test image to using `uploadtestimage.sh`. Defaults to $KO_DOCKER_REPO")

	fs.StringVar(&s.ImageTemplate, "env.image.template", "{{.Repository}}/{{.Name}}:{{.Tag}}",
		"Provide a template to generate the reference to an image from the test. Defaults to `{{.Repository}}/{{.Name}}:{{.Tag}}`.")

	fs.StringVar(&s.ImageTag, "env.image.tag", "latest", "Provide the version tag for the test images.")
}

func (s Settings) ImagePath(name string) string {
	tpl, err := template.New("image").Parse(s.ImageTemplate)
	if err != nil {
		panic("could not parse image template: " + err.Error())
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, struct {
		Repository string
		Name       string
		Tag        string
	}{
		Repository: s.ImageRepo,
		Name:       name,
		Tag:        s.ImageTag,
	}); err != nil {
		panic("could not apply the image template: " + err.Error())
	}
	return buf.String()
}

func (s Settings) ClientConfig() *rest.Config {
	overrides := &clientcmd.ConfigOverrides{}
	overrides.Context.Cluster = s.Cluster

	loader := &clientcmd.ClientConfigLoadingRules{ExplicitPath: s.KubeConfig}

	c, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides).ClientConfig()
	if err != nil {
		panic(err)
	}
	return c
}
