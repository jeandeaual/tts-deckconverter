package pkm

import (
	"fmt"
	"testing"
	"unicode"

	pokemontcgsdk "github.com/PokemonTCG/pokemon-tcg-sdk-go/src"
	"github.com/stretchr/testify/assert"
)

func TestBuildCost(t *testing.T) {
	assert.Equal(t, "", buildCost([]string{}))
	assert.Equal(
		t,
		"Colorless×2",
		buildCost([]string{"Colorless", "Colorless"}),
	)
	assert.Equal(
		t,
		"[00aaff]Water[ffffff]×1 [ff4040]Fire[ffffff]×2",
		buildCost([]string{"Water", "Fire", "Fire"}),
	)
	assert.Equal(
		t,
		"[cc00dd]Psychic[ffffff]×2 [333333]Darkness[ffffff]×1 Colorless×2",
		buildCost([]string{"Psychic", "Psychic", "Darkness", "Colorless", "Colorless"}),
	)
	assert.Equal(
		t,
		"[cc00dd]Psychic[ffffff]×1",
		buildCost([]string{"Psychic"}),
	)
	assert.Equal(
		t,
		"[cc00dd]Psychic[ffffff]×1 Colorless×1",
		buildCost([]string{"Psychic", "Colorless"}),
	)
}

func assertNoSpaceStartEnd(t *testing.T, description string) {
	runes := []rune(description)
	if len(runes) > 0 {
		assert.False(t, unicode.IsSpace(runes[0]), fmt.Sprintf("first letter shouldn't be a space: %s", description))
		assert.False(t, unicode.IsSpace(runes[len(runes)-1]), fmt.Sprintf("last letter shouldn't be a space: %s", description))
	}
}

func TestBuildCardDescription(t *testing.T) {
	assertNoSpaceStartEnd(t, buildCardDescription(pokemontcgsdk.PokemonCard{
		SuperType: "Pokémon",
		SubType:   "Basic",
		Text:      []string{"Test"},
	}))
	assertNoSpaceStartEnd(t, buildCardDescription(pokemontcgsdk.PokemonCard{
		SuperType: "Pokémon",
		SubType:   "Basic",
		Text:      []string{"Test 1", "Test 2"},
	}))
	assertNoSpaceStartEnd(t, buildCardDescription(pokemontcgsdk.PokemonCard{
		SuperType: "Pokémon",
		SubType:   "Basic",
		Text:      []string{"Test"},
		Attacks: []pokemontcgsdk.Attack{
			{
				Cost:   []string{"Psychic", "Colorless"},
				Name:   "Test",
				Damage: "50",
				Text:   "Test",
			},
			{
				Cost: []string{"Colorless"},
				Name: "Test",
			},
		},
	}))
}
