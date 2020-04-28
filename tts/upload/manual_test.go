package upload

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManualUpload(t *testing.T) {
	uploader := ManualUploader{}

	assert.Equal(t, "Manual", uploader.UploaderName())

	url, err := uploader.Upload("test.png", "Test", nil)
	assert.Nil(t, err)
	assert.Equal(t, "{{ test.png }}", url)
}
