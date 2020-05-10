package vanguard

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
	main, err := parseDeckFile(strings.NewReader(""))
	assert.Nil(t, main)
	assert.Nil(t, err)

	main, err = parseDeckFile(
		strings.NewReader(`Grade 3:
-4x Arboros Dragon, Sephirot
-3x Fruits Assort Dragon

Grade 2:
-4x Arboros Dragon, Timber
-4x Gale of Arboros, Oliver
-4x Spiritual Tree Sage, Irminsul
-2x Pansy Musketeer, Sylvia

Grade 1:
-4x Arboros Dragon, Branch
-4x Fruiting Wheat Maiden, Enifa
-4x Maiden of Happy Fawn

Grade 0:
-1x Arboros Dragon, Ratoon`),
	)
	expected := &CardNames{
		Names: []string{
			"Arboros Dragon, Sephirot",
			"Fruits Assort Dragon",
			"Arboros Dragon, Timber",
			"Gale of Arboros, Oliver",
			"Spiritual Tree Sage, Irminsul",
			"Pansy Musketeer, Sylvia",
			"Arboros Dragon, Branch",
			"Fruiting Wheat Maiden, Enifa",
			"Maiden of Happy Fawn",
			"Arboros Dragon, Ratoon",
		},
		Counts: map[string]int{
			"Arboros Dragon, Sephirot":      4,
			"Fruits Assort Dragon":          3,
			"Arboros Dragon, Timber":        4,
			"Gale of Arboros, Oliver":       4,
			"Spiritual Tree Sage, Irminsul": 4,
			"Pansy Musketeer, Sylvia":       2,
			"Arboros Dragon, Branch":        4,
			"Fruiting Wheat Maiden, Enifa":  4,
			"Maiden of Happy Fawn":          4,
			"Arboros Dragon, Ratoon":        1,
		},
	}

	assert.Equal(t, expected, main)
	assert.Nil(t, err)
}
