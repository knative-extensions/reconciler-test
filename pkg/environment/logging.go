package environment

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/multierr"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/system"

	"knative.dev/reconciler-test/pkg/k8s"
)

// logFunc type is a function that logs the given arguments
// according to the given format.
type logFunc func(format string, args ...interface{})

// exportLogs exports pods logs to an artifact directory
// as specified by the `ARTIFACTS` environment variable.
func (mr *MagicEnvironment) exportLogs() {
	logFunc := mr.logFunc()

	if isCI() {
		if err := exportLogs(mr.c); err != nil {
			logFunc("failed to export logs: %v", err)
		}
	}
}

// logFunc returns an environment specific logFunc.
func (mr *MagicEnvironment) logFunc() logFunc {
	var logFunc logFunc
	if mr.managedT != nil {
		logFunc = mr.managedT.Logf
	} else {
		logFunc = log.Printf
	}
	return logFunc
}

func exportLogs(ctx context.Context) error {

	namespace := FromContext(ctx).Namespace()

	// Create a directory for the namespace.
	logPath := filepath.Join(getLocalArtifactsDir(), system.Namespace(), namespace)
	if err := createDir(logPath); err != nil {
		return fmt.Errorf("error creating directory %q: %w", namespace, err)
	}

	logs, err := k8s.LogsForNamespace(kubeclient.Get(ctx), namespace)
	if err != nil && len(logs) == 0 {
		return err
	}

	globalErr := multierr.Combine(err)

	for podName, podLogs := range logs {
		func() {
			fileName := filepath.Join(logPath, podName)
			f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
			if err != nil {
				globalErr = multierr.Combine(globalErr, err)
				return
			}
			defer f.Close()

			_, _ = io.Copy(f, strings.NewReader(podLogs))
		}()
	}
	return globalErr
}

// isCI returns whether the current environment is a CI environment.
func isCI() bool {
	return strings.EqualFold(os.Getenv("CI"), "true")
}

// createDir creates dir if it does not exist.
// The created dir will have the permission bits as 0777,
// which means everyone can read/write/execute it.
func createDir(dirPath string) error {
	return createDirWithFileMode(dirPath, 0777)
}

// CreateDirWithFileMode creates dir if does not exist.
// The created dir will have the permission bits as perm, which is the standard Unix rwxrwxrwx permissions.
func createDirWithFileMode(dirPath string, perm os.FileMode) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err = os.MkdirAll(dirPath, perm); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}
	return nil
}

const (
	// artifactsDir is the default dir containing artifacts.
	// To change directory specify an ARTIFACTS environment
	// variable
	artifactsDir = "artifacts"
)

// getLocalArtifactsDir gets the artifact directory where
// prow looks for artifacts.
// By default, it will look at the env var ARTIFACTS.
func getLocalArtifactsDir() string {
	dir := os.Getenv("ARTIFACTS")
	if dir == "" {
		dir = artifactsDir
	}
	return dir
}
