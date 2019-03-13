package ygo

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"

	"deckconverter/plugins/ygo/api"
)

func assertNoSpaceStartEnd(t *testing.T, description string) {
	runes := []rune(description)
	assert.False(t, unicode.IsSpace(runes[0]), "first letter shouldn't be a space")
	assert.False(t, unicode.IsSpace(runes[len(runes)-1]), "last letter shouldn't be a space")
}

func TestBuildCardDescription(t *testing.T) {
	testText := "Test"
	assertNoSpaceStartEnd(t, buildDescription(api.Data{
		Name:        testText,
		Description: testText,
		Type:        api.TypeNormalMonster,
	}))
}
