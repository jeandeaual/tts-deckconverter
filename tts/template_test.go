package tts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindTemplateSize(t *testing.T) {
	col, row, err := findTemplateSize(1)
	assert.Equal(t, uint(1), col)
	assert.Equal(t, uint(1), row)
	assert.Nil(t, err)
	_, _, err = findTemplateSize(100)
	assert.NotNil(t, err)
	col, row, err = findTemplateSize(15)
	assert.Equal(t, uint(5), col)
	assert.Equal(t, uint(3), row)
	assert.Nil(t, err)
	col, row, err = findTemplateSize(16)
	assert.Equal(t, uint(4), col)
	assert.Equal(t, uint(4), row)
	assert.Nil(t, err)
	col, row, err = findTemplateSize(43)
	assert.Equal(t, uint(8), col)
	assert.Equal(t, uint(6), row)
	assert.Nil(t, err)
	col, row, err = findTemplateSize(50)
	assert.Equal(t, uint(8), col)
	assert.Equal(t, uint(7), row)
	assert.Nil(t, err)
	col, row, err = findTemplateSize(52)
	assert.Equal(t, uint(8), col)
	assert.Equal(t, uint(7), row)
	assert.Nil(t, err)
	col, row, err = findTemplateSize(60)
	assert.Equal(t, uint(9), col)
	assert.Equal(t, uint(7), row)
	assert.Nil(t, err)
	col, row, err = findTemplateSize(65)
	assert.Equal(t, uint(10), col)
	assert.Equal(t, uint(7), row)
	assert.Nil(t, err)
	col, row, err = findTemplateSize(70)
	assert.Equal(t, uint(10), col)
	assert.Equal(t, uint(7), row)
	assert.Nil(t, err)
}
