package plugins

import (
	"unicode"
)

func IndexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}

	// Not found.
	return -1
}

func CapitalizeString(s string) string {
	a := []rune(s)
	a[0] = unicode.ToUpper(a[0])
	return string(a)
}

func UncapitalizeString(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

func CapitalizeStrings(s []string) []string {
	new := make([]string, len(s))

	for i, el := range s {
		new[i] = CapitalizeString(el)
	}

	return new
}
