package ygo

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
	main, extra, side, err := parseDeckFile(strings.NewReader(""))
	assert.Nil(t, main)
	assert.Nil(t, extra)
	assert.Nil(t, side)
	assert.Nil(t, err)

	main, side, extra, err = parseDeckFile(
		strings.NewReader(`Main:

1 Leotron
1 Texchanger
1 Widget Kid
1 Cyberse White Hat
2 Bitron

Extra:

1 Transcode Talker
1 Pentestag`),
	)
	expected := &CardNames{
		Names: []string{
			"Leotron",
			"Texchanger",
			"Widget Kid",
			"Cyberse White Hat",
			"Bitron",
		},
		Counts: map[string]int{
			"Leotron":           1,
			"Texchanger":        1,
			"Widget Kid":        1,
			"Cyberse White Hat": 1,
			"Bitron":            2,
		},
	}

	assert.Equal(t, expected, main)

	expected = &CardNames{
		Names: []string{
			"Transcode Talker",
			"Pentestag",
		},
		Counts: map[string]int{
			"Transcode Talker": 1,
			"Pentestag":        1,
		},
	}

	assert.Equal(t, expected, side)

	assert.Nil(t, extra)
	assert.Nil(t, err)
}
