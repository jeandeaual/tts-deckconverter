package vanguard

import (
	"fmt"
	"testing"
	"unicode"

	"github.com/jeandeaual/tts-deckconverter/plugins/vanguard/cardfightwiki"
	"github.com/stretchr/testify/assert"
)

func assertNoSpaceStartEnd(t *testing.T, description string) {
	runes := []rune(description)
	if len(runes) > 0 {
		assert.False(t, unicode.IsSpace(runes[0]), fmt.Sprintf("first letter shouldn't be a space: %s", description))
		assert.False(t, unicode.IsSpace(runes[len(runes)-1]), fmt.Sprintf("last letter shouldn't be a space: %s", description))
	}
}

func TestBuildCardDescription(t *testing.T) {
	skill := "Twin Drive!!"
	power := "10000"
	nation := "United Sanctuary"
	assertNoSpaceStartEnd(t, buildCardDescription(cardfightwiki.Card{
		Grade:   3,
		Skill:   &skill,
		Power:   &power,
		Nation:  &nation,
		Clan:    "Royal Paladin",
		Race:    "Human",
		Formats: []string{"Premium Standard"},
	}))
}
