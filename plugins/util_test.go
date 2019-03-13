package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexOf(t *testing.T) {
	assert.Equal(t, 0, IndexOf("name", []string{"name", "description"}))
	assert.Equal(t, 3, IndexOf("", []string{"a", "b", "c", ""}))
	assert.Equal(t, -1, IndexOf("name", []string{}))
	assert.Equal(t, -1, IndexOf("", []string{"a", "b", "c"}))
}
