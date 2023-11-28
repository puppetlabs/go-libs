// Package sliceutil provides useful functions for working with slices not provided by the standard library.
package sliceutil

import (
	"reflect"
	"strings"
)

// RemoveZeroValues returns a copy of s with all zero-valued elements removed.
func RemoveZeroValues[S ~[]E, E any](s S) S {
	updatedSlice := make([]E, 0)
	for _, v := range s {
		if !reflect.ValueOf(v).IsZero() {
			updatedSlice = append(updatedSlice, v)
		}
	}

	return updatedSlice
}

// ToLower returns a copy of s with all letters in all elements mapped to lower case.
func ToLower(s []string) []string {
	lowerCasedSlice := make([]string, 0, len(s))
	for _, v := range s {
		lowerCased := strings.ToLower(v)
		lowerCasedSlice = append(lowerCasedSlice, lowerCased)
	}

	return lowerCasedSlice
}

// Trim returns a copy of s with all leading and trailing whitespace in all elements removed.
func Trim(s []string) []string {
	trimmedSlice := make([]string, 0, len(s))
	for _, v := range s {
		trimmed := strings.TrimSpace(v)
		trimmedSlice = append(trimmedSlice, trimmed)
	}

	return trimmedSlice
}
