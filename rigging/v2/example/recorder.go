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
		"knative.dev/reconciler-test/images/recorder",
	)
}

const recorderTemplate = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: recorder
  namespace: {{ .namespace }}
spec:
  replicas: 1
  selector:
    matchLabels: &labels
      app: recorder
  template:
    metadata:
      labels: *labels
    spec:
      serviceAccount: recorder
      containers:
        - name: recorder
          image: {{ .images.recorder }}
          env:
            - name: SYSTEM_NAMESPACE
              value: {{ .namespace }}
            - name: OBSERVER_NAME
              value: recorder-{{ .namespace }}
            - name: K8S_EVENT_SINK
              value: '{"apiVersion": "v1", "kind": "Namespace", "name": "{{ .namespace }}"}'

---

kind: Service
apiVersion: v1
metadata:
  name: recorder
  namespace: {{ .namespace }}
spec:
  selector:
    app: recorder
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: recorder
  namespace: {{ .namespace }}

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: event-watcher-{{ .namespace }}
rules:
  - apiGroups: [ "" ]
    resources:
      - "namespaces"
    verbs:
      - "get"
      - "list"
      - "watch"
  - apiGroups: [ "" ]
    resources:
      - "events"
    verbs:
      - "get"
      - "list"
      - "create"
      - "update"
      - "delete"
      - "patch"
      - "watch"

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: event-watcher-{{ .namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: event-watcher-{{ .namespace }}
subjects:
  - kind: ServiceAccount
    name: recorder
    namespace: {{ .namespace }}
`

func InstallRecorder(t riggingv2.PT, e riggingv2.Environment) {
	// Install the producer.

	// TODO: move this next to the config, and use ParseTemplates to implement this method.
}
