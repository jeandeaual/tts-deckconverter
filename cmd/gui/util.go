package main

import (
	"unicode"
)

func capitalizeString(s string) string {
	a := []rune(s)
	a[0] = unicode.ToUpper(a[0])
	return string(a)
}

func uncapitalizeString(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

func capitalizeStrings(s []string) []string {
	new := make([]string, len(s))

	for i, el := range s {
		new[i] = capitalizeString(el)
	}

	return new
}
