package k8s

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"strings"
)

func LogsFor(client kubernetes.Interface, namespace, name string, gvr schema.GroupVersionResource) (string, error) {
	// Get all pods in this namespace.
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	logs := make([]string, 0)

	// Look for a pod with the name that was passed in inside the pod name.
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, name) {
			// Collect all the logs from all the containers for this pod.
			if l, err := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{}).DoRaw(context.Background()); err != nil {
				logs = append(logs, err.Error())
			} else {
				logs = append(logs, string(l))
			}
		}
	}

	// Did we find a match like the given name?
	if len(logs) == 0 {
		return "", fmt.Errorf(`pod for "%s/%s" [%s] not found`, namespace, name, gvr.String())
	}

	return strings.Join(logs, "\n"), nil
}
