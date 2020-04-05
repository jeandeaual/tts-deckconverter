package pkm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func init() {
	logger := zap.NewExample()
	log = logger.Sugar()
}

func TestSetup(t *testing.T) {
	ok := setUp(log)
	assert.True(t, ok)
}

func TestGetSetCode(t *testing.T) {
	set, ok := getSetCode("BS", log)
	assert.True(t, ok)
	assert.Equal(t, "base1", set)

	set, ok = getSetCode("LOT", log)
	assert.True(t, ok)
	assert.Equal(t, "sm8", set)

	set, ok = getSetCode("DRM", log)
	assert.True(t, ok)
	assert.Equal(t, "sm75", set)

	set, ok = getSetCode("UNB", log)
	assert.True(t, ok)
	assert.Equal(t, "sm10", set)

	_, ok = getSetCode("INVALID", log)
	assert.False(t, ok)
}

func TestGetPTCGOSetCode(t *testing.T) {
	set, ok := getPTCGOSetCode("base1", log)
	assert.True(t, ok)
	assert.Equal(t, "BS", set)

	set, ok = getPTCGOSetCode("SM8", log)
	assert.True(t, ok)
	assert.Equal(t, "LOT", set)

	set, ok = getPTCGOSetCode("SM75", log)
	assert.True(t, ok)
	assert.Equal(t, "DRM", set)

	set, ok = getPTCGOSetCode("SM10", log)
	assert.True(t, ok)
	assert.Equal(t, "UNB", set)

	_, ok = getPTCGOSetCode("INVALID", log)
	assert.False(t, ok)
}
