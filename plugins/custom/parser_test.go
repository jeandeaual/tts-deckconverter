package custom

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
	deck, err := parseList(strings.NewReader(""))
	assert.Nil(t, deck)
	assert.Nil(t, err)

	deck, err = parseList(
		strings.NewReader(`// Deck
1 https://img.scryfall.com/cards/large/front/7/3/732fa4c9-11da-4bdb-96af-aa37c74be25f.jpg?1562803341
https://img.scryfall.com/cards/large/front/f/3/f3dd4d92-6471-4f4b-9c70-cbd2196e8c7b.jpg?1562640130 (Horse)
/home/user/test.jpg (Test 1)
3 C:\Users\User\Test\Test Card.jpg (Test 2)`),
	)
	expectedNames := []string{
		"Horse",
		"Test 1",
		"Test 2",
	}
	expected := &CardFiles{
		Cards: []CardInfo{
			{
				Path: "https://img.scryfall.com/cards/large/front/7/3/732fa4c9-11da-4bdb-96af-aa37c74be25f.jpg?1562803341",
				Name: nil,
			},
			{
				Path: "https://img.scryfall.com/cards/large/front/f/3/f3dd4d92-6471-4f4b-9c70-cbd2196e8c7b.jpg?1562640130",
				Name: &expectedNames[0],
			},
			{
				Path: "/home/user/test.jpg",
				Name: &expectedNames[1],
			},
			{
				Path: "C:\\Users\\User\\Test\\Test Card.jpg",
				Name: &expectedNames[2],
			},
		},
		Counts: map[string]int{
			"https://img.scryfall.com/cards/large/front/7/3/732fa4c9-11da-4bdb-96af-aa37c74be25f.jpg?1562803341": 1,
			"https://img.scryfall.com/cards/large/front/f/3/f3dd4d92-6471-4f4b-9c70-cbd2196e8c7b.jpg?1562640130": 1,
			"/home/user/test.jpg":                  1,
			"C:\\Users\\User\\Test\\Test Card.jpg": 3,
		},
	}

	assert.Equal(t, expected, deck)
	assert.Nil(t, err)
}
