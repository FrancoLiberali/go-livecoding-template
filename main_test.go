package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Table-driven test scaffold with testify — copy this shape per function.
// assert.* keeps going on failure; require.* stops the test (use for
// preconditions like require.NoError before asserting the value).
func TestExample(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{name: "zero", in: 0, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := identity(tt.in) // replace with the function under test
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// placeholder under test — delete when you write the real thing.
func identity(n int) (int, error) { return n, nil }
