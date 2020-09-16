module knative.dev/reconciler-test

go 1.14

require (
	github.com/onsi/gomega v1.10.2
	k8s.io/apimachinery v0.18.8
	knative.dev/pkg v0.0.0-20200916171541-6e0430fd94db
	knative.dev/test-infra v0.0.0-20200911201000-3f90e7c8f2fa
)

replace (
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
)
