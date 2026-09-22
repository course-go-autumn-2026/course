package linterdemo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePort(t *testing.T) {
	port, err := ParsePort("8080")
	require.NoError(t, err)
	assert.Equal(t, 8080, port)

	_, err = ParsePort("0")
	assert.True(t, errors.Is(err, ErrInvalidPort))
}
