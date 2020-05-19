package upload

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManualUpload(t *testing.T) {
	workingDir, err := os.Getwd()
	assert.Nil(t, err)

	uploader := ManualUploader{}

	assert.Equal(t, "Manual", uploader.UploaderName())

	url, err := uploader.Upload("test.png", "Test", nil)
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(workingDir, "test.png"), url)
}
