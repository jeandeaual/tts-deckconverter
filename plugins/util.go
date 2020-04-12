package plugins

import (
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
