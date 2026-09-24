package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGreet(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "no args", args: nil, want: "hello, world"},
		{name: "one arg", args: []string{"alice"}, want: "hello, alice"},
		{name: "many args", args: []string{"alice", "bob"}, want: "hello, alice bob"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, greet(tt.args))
		})
	}
}

func TestRun(t *testing.T) {
	var buf bytes.Buffer

	require.NoError(t, run([]string{"bob"}, &buf))
	assert.Equal(t, "hello, bob\n", buf.String())
}
