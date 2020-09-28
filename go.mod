module knative.dev/reconciler-test

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/octago/sflags v0.2.0
	github.com/onsi/gomega v1.10.2
	go.opencensus.io v0.22.4
	golang.org/x/tools v0.0.0-20200918201133-e94ab7288189 // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
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
