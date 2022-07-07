/*
Copyright 2022 The Knative Authors

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

package manifest_test

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"os"
	"testing"

	"gotest.tools/v3/assert"
	testlog "knative.dev/reconciler-test/pkg/logging"
	"knative.dev/reconciler-test/pkg/manifest"
)

//go:embed testdata/*.yaml
var templates embed.FS

func TestParseTemplatesFS(t *testing.T) {
	ns := randomString(t, 10)
	ctx := testlog.WithTestLogger(context.TODO(), t)
	data := map[string]interface{}{
		"namespace": ns,
	}
	images := map[string]string{}
	yamlsDir, err := manifest.ParseTemplatesFS(ctx, templates, images, data)
	assert.NilError(t, err)
	fileInfo, err := os.Stat(yamlsDir)
	assert.NilError(t, err)
	assert.Assert(t, fileInfo.IsDir())
	err = os.RemoveAll(yamlsDir)
	assert.NilError(t, err)
}

func randomString(t fatalt, length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		t.Fatal(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

type fatalt interface {
	Fatal(...interface{})
}
