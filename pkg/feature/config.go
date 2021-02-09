package feature

import (
	"fmt"

	"knative.dev/reconciler-test/pkg/config"
)

func (f *Feature) Config(out interface{}) interface{} {
	err := config.UnmarshalConfig(f.Name, out)
	if err != nil {
		panic(fmt.Sprintf("cannot read config: %v", err))
	}

	return out
}
