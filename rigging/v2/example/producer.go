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

package example

import (
	"knative.dev/reconciler-test/rigging"
	riggingv2 "knative.dev/reconciler-test/rigging/v2"
)

func init() {
	rigging.RegisterPackage(
		"knative.dev/reconciler-test/rigging/v2/example/cmd/producer",
	)
}

const producerTemplate = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: producer
  namespace: {{ .namespace }}
spec:
  replicas: 1
  selector:
    matchLabels: &labels
      app: producer
  template:
    metadata:
      labels: *labels
    spec:
      containers:
        - name: producer
          image: {{ .images.producer }}
          env:
            - name: COUNT
              value: '{{ .producerCount }}'
            - name: K_SINK
              # TODO: was trying to inject the namespace but it is unknown at the time this is passed.
              value: 'http://{{ .producerSink }}.{{ .namespace }}.svc.{{ .clusterDomainName }}'
`

func InstallProducer(t riggingv2.PT, e riggingv2.Environment) {
	// Install the producer.

}
