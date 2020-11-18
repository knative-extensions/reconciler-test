module knative.dev/reconciler-test

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/google/uuid v1.1.2
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/onsi/gomega v1.10.2 // indirect
	github.com/pkg/errors v0.9.1
	go.opencensus.io v0.22.5
	go.uber.org/zap v1.16.0
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/tools v0.0.0-20200918201133-e94ab7288189 // indirect
	k8s.io/api v0.18.12
	k8s.io/apimachinery v0.18.12
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/klog v1.0.0
	knative.dev/hack v0.0.0-20201118155651-b31d3bb6bff9
	knative.dev/pkg v0.0.0-20201117221452-0fccc54273ed
)

replace (
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
)
