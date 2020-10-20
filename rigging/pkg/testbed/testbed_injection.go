package testbed

import (
	"context"

	"knative.dev/pkg/injection/clients/dynamicclient"
)

func New(ctx context.Context) TestBed {
	return &testBed{
		client: dynamicclient.Get(ctx),
	}
}
