module knative.dev/reconciler-test

go 1.16

require (
	github.com/cloudevents/conformance v0.2.0
	github.com/cloudevents/sdk-go/v2 v2.4.1
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-openapi/spec v0.19.6 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/onsi/gomega v1.10.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	go.opencensus.io v0.23.0
	go.uber.org/atomic v1.9.0
	go.uber.org/zap v1.19.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	k8s.io/api v0.21.4
	k8s.io/apimachinery v0.21.4
	k8s.io/client-go v0.21.4
	k8s.io/klog v1.0.0
	knative.dev/hack v0.0.0-20211101195839-11d193bf617b
	knative.dev/pkg v0.0.0-20211101212339-96c0204a70dc
)
