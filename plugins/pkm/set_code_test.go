package pkm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"deckconverter/log"
)

func init() {
	logger := zap.NewExample()
	log.SetLogger(logger.Sugar())
}

func TestSetup(t *testing.T) {
	ok := setUp()
	assert.True(t, ok)
}

func TestGetSetCode(t *testing.T) {
	set, ok := getSetCode("BS")
	assert.True(t, ok)
	assert.Equal(t, "base1", set)

	set, ok = getSetCode("LOT")
	assert.True(t, ok)
	assert.Equal(t, "sm8", set)

	set, ok = getSetCode("DRM")
	assert.True(t, ok)
	assert.Equal(t, "sm75", set)

	set, ok = getSetCode("UNB")
	assert.True(t, ok)
	assert.Equal(t, "sm10", set)

	_, ok = getSetCode("INVALID")
	assert.False(t, ok)
}

func TestGetPTCGOSetCode(t *testing.T) {
	set, ok := getPTCGOSetCode("base1")
	assert.True(t, ok)
	assert.Equal(t, "BS", set)

	set, ok = getPTCGOSetCode("SM8")
	assert.True(t, ok)
	assert.Equal(t, "LOT", set)

	set, ok = getPTCGOSetCode("SM75")
	assert.True(t, ok)
	assert.Equal(t, "DRM", set)

	set, ok = getPTCGOSetCode("SM10")
	assert.True(t, ok)
	assert.Equal(t, "UNB", set)

	_, ok = getPTCGOSetCode("INVALID")
	assert.False(t, ok)
}
