package pkm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDeckFile(t *testing.T) {
	main, err := parseDeckFile(strings.NewReader(""))
	assert.Nil(t, main)
	assert.Nil(t, err)

	main, err = parseDeckFile(
		strings.NewReader(`##Pokémon - 10

* 2 Furfrou KSS 32
* 2 Hoothoot PLF 91
* 2 Miltank FLF 83
* 3 Skitty XY 104
* 1 Delcatty XY 105

##Trainer Cards - 6

* 2 Escape Rope PLS 120
* 2 Roller Skates XY 125
* 1 Pokémon Center Lady FLF 93
* 1 Hard Charm XY 119

##Energy - 18

* 18 Psychic Energy XYEnergy 8

Total Cards - 34
`),
	)
	expected := &CardNames{
		Names: []CardInfo{
			{
				Name:   "Furfrou",
				Set:    "KSS",
				Number: "32",
			},
			{
				Name:   "Hoothoot",
				Set:    "PLF",
				Number: "91",
			},
			{
				Name:   "Miltank",
				Set:    "FLF",
				Number: "83",
			},
			{
				Name:   "Skitty",
				Set:    "XY",
				Number: "104",
			},
			{
				Name:   "Delcatty",
				Set:    "XY",
				Number: "105",
			},
			{
				Name:   "Escape Rope",
				Set:    "PLS",
				Number: "120",
			},
			{
				Name:   "Roller Skates",
				Set:    "XY",
				Number: "125",
			},
			{
				Name:   "Pokémon Center Lady",
				Set:    "FLF",
				Number: "93",
			},
			{
				Name:   "Hard Charm",
				Set:    "XY",
				Number: "119",
			},
			{
				Name:   "Psychic Energy",
				Set:    "XYEnergy",
				Number: "8",
			},
		},
		Counts: map[string]int{
			"Furfrou":             2,
			"Hoothoot":            2,
			"Miltank":             2,
			"Skitty":              3,
			"Delcatty":            1,
			"Escape Rope":         2,
			"Roller Skates":       2,
			"Pokémon Center Lady": 1,
			"Hard Charm":          1,
			"Psychic Energy":      18,
		},
	}

	assert.Equal(t, expected, main)
	assert.Nil(t, err)
}
