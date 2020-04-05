package pkm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDeckFile(t *testing.T) {
	main, err := parseDeckFile(strings.NewReader(""), log)
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
		log,
	)
	expected := &CardNames{
		Names: []CardInfo{
			CardInfo{
				Name:   "Furfrou",
				Set:    "KSS",
				Number: "32",
			},
			CardInfo{
				Name:   "Hoothoot",
				Set:    "PLF",
				Number: "91",
			},
			CardInfo{
				Name:   "Miltank",
				Set:    "FLF",
				Number: "83",
			},
			CardInfo{
				Name:   "Skitty",
				Set:    "XY",
				Number: "104",
			},
			CardInfo{
				Name:   "Delcatty",
				Set:    "XY",
				Number: "105",
			},
			CardInfo{
				Name:   "Escape Rope",
				Set:    "PLS",
				Number: "120",
			},
			CardInfo{
				Name:   "Roller Skates",
				Set:    "XY",
				Number: "125",
			},
			CardInfo{
				Name:   "Pokémon Center Lady",
				Set:    "FLF",
				Number: "93",
			},
			CardInfo{
				Name:   "Hard Charm",
				Set:    "XY",
				Number: "119",
			},
			CardInfo{
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
