package main

import (
	"fmt"

	"knative.dev/pkg/injection"

	"knative.dev/reconciler-test/rigging/pkg/testbed"
)

func main() {
	ctx, _ := injection.EnableInjectionOrDie(nil, nil)

	// Start the testbed.
	tb := testbed.New(ctx)
	if err := tb.Start(ctx); err != nil {
		panic(fmt.Sprintf("Failed running testbed - %s", err))
	}
}
