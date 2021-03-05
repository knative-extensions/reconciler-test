package feature

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFeature(t *testing.T) {
	f := NewFeature()
	require.Equal(t, "TestNewFeature", f.Name)
}
