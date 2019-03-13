package magic

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func init() {
	logger := zap.NewExample()
	log = logger.Sugar()
}

func TestParseDeckFile(t *testing.T) {
	main, side, err := parseDeckFile(strings.NewReader(""), log)
	assert.Nil(t, main)
	assert.Nil(t, side)
	assert.Nil(t, err)

	setRNA := "RNA"
	setM19 := "M19"
	setDOM := "DOM"
	setGRN := "GRN"
	setXLN := "XLN"
	setEMN := "EMN"
	main, side, err = parseDeckFile(
		strings.NewReader(`2 Blood Crypt (RNA) 245
3 Carnival /// Carnage (RNA) 222
3 Demon of Catastrophes (M19) 91
1 Demonlord Belzenlok (DOM) 86
4 Diregraf Ghoul (M19) 92
1 Doom Whisperer (GRN) 69
4 Dragonskull Summit (XLN) 252
2 Graf Rats (EMN) 91a
2 Midnight Scavengers (EMN) 96a`),
		log,
	)
	expected := &CardNames{
		Names: []CardInfo{
			CardInfo{
				Name: "Blood Crypt",
				Set:  &setRNA,
			},
			CardInfo{
				Name: "Carnival // Carnage",
				Set:  &setRNA,
			},
			CardInfo{
				Name: "Demon of Catastrophes",
				Set:  &setM19,
			},
			CardInfo{
				Name: "Demonlord Belzenlok",
				Set:  &setDOM,
			},
			CardInfo{
				Name: "Diregraf Ghoul",
				Set:  &setM19,
			},
			CardInfo{
				Name: "Doom Whisperer",
				Set:  &setGRN,
			},
			CardInfo{
				Name: "Dragonskull Summit",
				Set:  &setXLN,
			},
			CardInfo{
				Name: "Graf Rats",
				Set:  &setEMN,
			},
			CardInfo{
				Name: "Midnight Scavengers",
				Set:  &setEMN,
			},
		},
		Counts: map[string]int{
			"Blood Crypt":           2,
			"Carnival // Carnage":   3,
			"Demon of Catastrophes": 3,
			"Demonlord Belzenlok":   1,
			"Diregraf Ghoul":        4,
			"Doom Whisperer":        1,
			"Dragonskull Summit":    4,
			"Graf Rats":             2,
			"Midnight Scavengers":   2,
		},
	}

	assert.Equal(t, expected, main)
	assert.Nil(t, side)
	assert.Nil(t, err)
}
