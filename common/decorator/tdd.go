// Package decorator is TDD basic implement, All test cases must be tested using the test suite
package decorator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Case[T any] struct {
	Name    string
	Input   T
	Want    any
	WantErr string
	Skip    string
	Setup   func(t *testing.T)
	Cleanup func(t *testing.T)
}

func RunCases[T any](t *testing.T, cases []Case[T], fn func(t *testing.T, input T) (any, error)) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Skip != "" {
				t.Skip(c.Skip)
			}
			if c.Setup != nil {
				c.Setup(t)
			}
			if c.Cleanup != nil {
				t.Cleanup(func() { c.Cleanup(t) })
			}

			got, err := fn(t, c.Input)

			if c.WantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), c.WantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, c.Want, got)
		})
	}
}
