module knative.dev/reconciler-test

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/google/uuid v1.1.2
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/onsi/gomega v1.10.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	go.opencensus.io v0.22.5
	go.uber.org/zap v1.16.0
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v0.19.7
	k8s.io/klog v1.0.0
	knative.dev/hack v0.0.0-20210114150620-4422dcadb3c8
	knative.dev/pkg v0.0.0-20210114223020-f0ea5e6b9c4e
)
