package plugins

import (
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

// IndexOf returns the index of a string in a string slice, or -1 if not found.
func IndexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}

	// Not found.
	return -1
}

// CapitalizeString puts the first letter of a string in uppercase.
func CapitalizeString(s string) string {
	a := []rune(s)
	a[0] = unicode.ToUpper(a[0])
	return string(a)
}

// UncapitalizeString puts the first letter of a string in lowercase.
func UncapitalizeString(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

// CapitalizeStrings puts the first letter of each string in a slice
// in uppercase.
func CapitalizeStrings(s []string) []string {
	new := make([]string, len(s))

	for i, el := range s {
		new[i] = CapitalizeString(el)
	}

	return new
}

// CheckInvalidFolderName checks whether the given path can be safely created
// on all platforms.
// Since the Tabletop Simulator save files can be shared with Steam Cloud on
// Windows, make sure that even the folders created on Linux or OS X are also
// valid on Windows.
func CheckInvalidFolderName(folderPath string) bool {
	if runtime.GOOS == "windows" {
		if filepath.IsAbs(folderPath) {
			// Ignore the drive letter
			folderPath = folderPath[2:]
		}

		return strings.ContainsAny(folderPath, "/:*?<>|")
	}

	return strings.ContainsAny(folderPath, "\\:*?<>|")
}
