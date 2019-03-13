package magic

import (
	"testing"
	"unicode"

	scryfall "github.com/BlueMonday/go-scryfall"
	"github.com/stretchr/testify/assert"
)

func assertNoSpaceStartEnd(t *testing.T, description string) {
	runes := []rune(description)
	assert.False(t, unicode.IsSpace(runes[0]), "first letter shouldn't be a space")
	assert.False(t, unicode.IsSpace(runes[len(runes)-1]), "last letter shouldn't be a space")
}

func TestBuildCardDescription(t *testing.T) {
	testText := "Test"
	assertNoSpaceStartEnd(t, buildCardDescription(scryfall.Card{
		Name:       testText,
		TypeLine:   testText,
		OracleText: testText,
	}, nil))
	assertNoSpaceStartEnd(t, buildCardDescription(scryfall.Card{
		Name:       testText,
		TypeLine:   testText,
		OracleText: testText,
		Power:      &testText,
		Toughness:  &testText,
	}, nil))
	assertNoSpaceStartEnd(t, buildCardDescription(scryfall.Card{
		Name:       testText,
		TypeLine:   testText,
		OracleText: testText,
		FlavorText: &testText,
		Loyalty:    &testText,
	}, nil))
}

func TestBuildCardFacesDescription(t *testing.T) {
	testText := "Test"
	assertNoSpaceStartEnd(t, buildCardFacesDescription([]scryfall.CardFace{
		scryfall.CardFace{
			Name:       testText,
			TypeLine:   testText,
			OracleText: &testText,
		},
		scryfall.CardFace{
			Name:       testText,
			TypeLine:   testText,
			OracleText: &testText,
			Power:      &testText,
			Toughness:  &testText,
		},
		scryfall.CardFace{
			Name:       testText,
			TypeLine:   testText,
			OracleText: &testText,
			FlavorText: &testText,
			Power:      &testText,
			Toughness:  &testText,
		},
		scryfall.CardFace{
			Name:       testText,
			TypeLine:   testText,
			OracleText: &testText,
			FlavorText: &testText,
			Loyalty:    &testText,
		},
	}, nil))
}
