package plugins

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexOf(t *testing.T) {
	assert.Equal(t, 0, IndexOf("name", []string{"name", "description"}))
	assert.Equal(t, 3, IndexOf("", []string{"a", "b", "c", ""}))
	assert.Equal(t, -1, IndexOf("name", []string{}))
	assert.Equal(t, -1, IndexOf("", []string{"a", "b", "c"}))
}

func TestCapitalizeString(t *testing.T) {
	assert.Equal(t, "Test", CapitalizeString("test"))
	assert.Equal(t, "TEST", CapitalizeString("TEST"))
	assert.Equal(t, "Αβ", CapitalizeString("αβ"))
	assert.Equal(t, "Авгдєѕзиѳ", CapitalizeString("авгдєѕзиѳ"))
}

func TestUncapitalizeString(t *testing.T) {
	assert.Equal(t, "test", UncapitalizeString("Test"))
	assert.Equal(t, "tEST", UncapitalizeString("TEST"))
	assert.Equal(t, "αβ", UncapitalizeString("Αβ"))
	assert.Equal(t, "аВГДЄЅЗИѲ", UncapitalizeString("АВГДЄЅЗИѲ"))
}

func TestCapitalizeStrings(t *testing.T) {
	assert.Equal(t, []string{"Test", "12345", "Hello"}, CapitalizeStrings([]string{"test", "12345", "Hello"}))
}

func TestCheckInvalidFolderName(t *testing.T) {
	if runtime.GOOS == "windows" {
		assert.Equal(t, false, CheckInvalidFolderName("C:\\Program Files\\Users\\Test"))
		assert.Equal(t, true, CheckInvalidFolderName("C:\\Program Files\\Users\\Hello/Test"))
		assert.Equal(t, true, CheckInvalidFolderName("/home/test"))
		assert.Equal(t, true, CheckInvalidFolderName("C:\\Program Files\\Users\\:"))
	} else {
		assert.Equal(t, false, CheckInvalidFolderName("/home/test"))
		assert.Equal(t, true, CheckInvalidFolderName("/home/test/tts\\folder"))
		assert.Equal(t, true, CheckInvalidFolderName("/home/test/tts: folder"))
	}
}
