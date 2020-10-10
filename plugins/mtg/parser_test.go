package mtg

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/jeandeaual/tts-deckconverter/log"
)

func init() {
	logger := zap.NewExample()
	log.SetLogger(logger.Sugar())
}

func TestParseDeckFile(t *testing.T) {
	main, side, maybe, err := parseDeckFile(strings.NewReader(""))
	assert.Nil(t, main)
	assert.Nil(t, side)
	assert.Nil(t, maybe)
	assert.Nil(t, err)

	setRNA := "RNA"
	setM19 := "M19"
	setDOM := "DOM"
	setGRN := "GRN"
	setXLN := "XLN"
	setEMN := "EMN"
	main, side, maybe, err = parseDeckFile(
		strings.NewReader(`2 Blood Crypt (RNA) 245
3 Carnival /// Carnage (RNA) 222
3 Demon of Catastrophes (M19) 91
1 Demonlord Belzenlok (DOM) 86
4 Diregraf Ghoul (M19) 92
1 Doom Whisperer (GRN) 69
4 Dragonskull Summit (XLN) 252
2 Graf Rats (EMN) 91a
2 Midnight Scavengers (EMN) 96a`),
	)
	expected := &CardNames{
		Names: []CardInfo{
			{
				Name: "Blood Crypt",
				Set:  &setRNA,
			},
			{
				Name: "Carnival // Carnage",
				Set:  &setRNA,
			},
			{
				Name: "Demon of Catastrophes",
				Set:  &setM19,
			},
			{
				Name: "Demonlord Belzenlok",
				Set:  &setDOM,
			},
			{
				Name: "Diregraf Ghoul",
				Set:  &setM19,
			},
			{
				Name: "Doom Whisperer",
				Set:  &setGRN,
			},
			{
				Name: "Dragonskull Summit",
				Set:  &setXLN,
			},
			{
				Name: "Graf Rats",
				Set:  &setEMN,
			},
			{
				Name: "Midnight Scavengers",
				Set:  &setEMN,
			},
		},
		Counts: map[string]int{
			"Blood CryptRNA":           2,
			"Carnival // CarnageRNA":   3,
			"Demon of CatastrophesM19": 3,
			"Demonlord BelzenlokDOM":   1,
			"Diregraf GhoulM19":        4,
			"Doom WhispererGRN":        1,
			"Dragonskull SummitXLN":    4,
			"Graf RatsEMN":             2,
			"Midnight ScavengersEMN":   2,
		},
	}

	assert.Equal(t, expected, main)
	assert.Nil(t, side)
	assert.Nil(t, maybe)
	assert.Nil(t, err)
}
