package sliceutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveZeroValues(t *testing.T) {
	boolSlice := []bool{false, false, false}
	require.Equal(t, []bool{}, RemoveZeroValues(boolSlice))

	intSlice := []int{1, 987, 0}
	require.Equal(t, []int{1, 987}, RemoveZeroValues(intSlice))

	strSlice := []string{"hello", "world", ""}
	require.Equal(t, []string{"hello", "world"}, RemoveZeroValues(strSlice))

	str1 := "hello"
	str2 := "world"
	str1Ptr := &str1
	str2Ptr := &str2
	var nilStrPtr *string

	strPtrSlice := []*string{&str1, &str2, nilStrPtr, nil}
	require.Equal(t, []*string{str1Ptr, str2Ptr}, RemoveZeroValues(strPtrSlice))

	type testStruct struct {
		flag bool
	}

	ts := []testStruct{{flag: true}, {flag: false}}
	require.Equal(t, []testStruct{{flag: true}}, RemoveZeroValues(ts))
}

func TestToLower(t *testing.T) {
	s := []string{"apples", "ORANGES", "**Pears**", "bAnAnAs-123"}
	require.Equal(t, []string{"apples", "oranges", "**pears**", "bananas-123"}, ToLower(s))
}

func TestTrim(t *testing.T) {
	s := []string{"no-space", "middle space", " left-space", "right-space ", "   surrounding-space   "}
	require.Equal(t, []string{"no-space", "middle space", "left-space", "right-space", "surrounding-space"}, Trim(s))
}
