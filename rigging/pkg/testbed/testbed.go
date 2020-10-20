package testbed

import (
	"context"
	"fmt"

	"k8s.io/client-go/dynamic"
	"knative.dev/reconciler-test/rigging/pkg/binding/manifest"
)

type testBed struct {
	client dynamic.Interface
}

func (b *testBed) Start(ctx context.Context) error {

	mf, err := manifest.New(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d resources.\n", len(mf.ResourceNames()))

	for _, n := range mf.ResourceNames() {
		fmt.Printf("Going to make %s\n", n)
	}

	if err := mf.ApplyAll(); err != nil {
		fmt.Printf("failed to apply all, %s\n", err)
		return err
	}

	fmt.Printf("wait forever...!\n")
	<-ctx.Done()
	fmt.Printf("done!\n")
	return nil
}
