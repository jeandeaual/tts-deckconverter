package ygo

import (
	"fmt"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"

	"github.com/jeandeaual/tts-deckconverter/plugins/ygo/api"
)

func assertNoSpaceStartEnd(t *testing.T, description string) {
	runes := []rune(description)
	if len(runes) > 0 {
		assert.False(t, unicode.IsSpace(runes[0]), fmt.Sprintf("first letter shouldn't be a space: %s", description))
		assert.False(t, unicode.IsSpace(runes[len(runes)-1]), fmt.Sprintf("last letter shouldn't be a space: %s", description))
	}
}

func TestBuildCardDescription(t *testing.T) {
	testText := "Test"
	assertNoSpaceStartEnd(t, buildDescription(api.Data{
		Name:        testText,
		Description: testText,
		Type:        api.TypeNormalMonster,
	}))
}
