package upload

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"go.uber.org/zap"

	"github.com/jeandeaual/tts-deckconverter/log"
)

func init() {
	logger := zap.NewExample()
	log.SetLogger(logger.Sugar())
}

// Create a temporary file of size bytes and return its path
func createTempFile(t *testing.T, size int64) string {
	// Create a temporary file
	tmpFile, err := ioutil.TempFile(os.TempDir(), "prefix-")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}
	defer func() {
		cerr := tmpFile.Close()
		if cerr != nil && !errors.Is(cerr, os.ErrClosed) {
			log.Fatal("Couldn't close temporary file %s: %v", tmpFile.Name(), cerr)
		}
	}()

	if size <= 0 {
		return tmpFile.Name()
	}

	// Fill with size bytes
	_, err = tmpFile.Seek(size-1, 0)
	if err != nil {
		log.Fatalf("Failed to seek to %d bytes", size)
	}
	_, err = tmpFile.Write([]byte{0})
	if err != nil {
		log.Fatalf("Write to %s failed", tmpFile.Name())
	}
	err = tmpFile.Close()
	if err != nil {
		log.Fatalf("Failed to close %s", tmpFile.Name())
	}

	t.Logf("Created temporary file %s (size: %d bytes)", tmpFile.Name(), size)

	return tmpFile.Name()
}

func removeFile(file string) {
	err := os.Remove(file)
	if err != nil {
		log.Fatal("Couldn't delete %s: %v", file, err)
	}
}
