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

package ko

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrKoPublishFailed is returned when the ko publish command fails.
var ErrKoPublishFailed = errors.New("ko publish failed")

// Publish uses ko to publish the image.
func Publish(ctx context.Context, path string) (string, error) {
	koPack := fmt.Sprintf("ko://%s", path)
	version := os.Getenv("GOOGLE_KO_VERSION")
	if version == "" {
		version = "v0.11.2"
	}
	args := []string{
		"go", "run", fmt.Sprintf("github.com/google/ko@%s", version),
		"publish",
	}
	platform := os.Getenv("PLATFORM")
	if len(platform) > 0 {
		args = append(args, "--platform="+platform)
	}
	args = append(args, "-B", koPack)
	out, err := runCmd(ctx, args)
	if err != nil {
		// Build errors are caught by build tests, so in case KO fails it might be because
		// `path` it's an already built image like "gcr.io/knative-nightly/knative.dev/eventing/cmd/heartbeats"
		if isGoPkg, pkgErr := isGoPackage(path); pkgErr != nil || isGoPkg {
			return "", fmt.Errorf("%w: %v -- command: %q, pkgErr %v",
				ErrKoPublishFailed, err, args, pkgErr)
		}
		return path, nil
	}
	return out, nil
}

func isGoPackage(path string) (bool, error) {
	/*
			Ideally, we would use:
				bi, ok := debug.ReadBuildInfo()
				if !ok {
					return false, fmt.Errorf("failed to read build info")
				}
				modFile, err := os.ReadFile(bi.Main.Path)
				if err != nil {
					return true, err
				}
				mp := modfile.ModulePath(modFile)

				return strings.Contains(path, mp), nil

			but due to https://github.com/golang/go/issues/33976, we cannot, so workaround it by
		    assuming that in Knative we will build modules that starts with knative.dev like
		    "knative.dev/reconciler-test/cmd/eventshub"
	*/
	return strings.HasPrefix(path, "knative.dev"), nil
}
