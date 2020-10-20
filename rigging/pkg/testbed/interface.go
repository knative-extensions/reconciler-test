package testbed

import "context"

type TestBed interface {
	Start(context.Context) error
}
