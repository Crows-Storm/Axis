package adapters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafePrefixStripsWildcards(t *testing.T) {
	assert.Equal(t, "ab", safePrefix(" a%b_ "))
	assert.Equal(t, "", safePrefix("   "))
}
