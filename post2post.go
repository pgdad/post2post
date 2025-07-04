package post2post

import (
	"fmt"
	"strings"
)

// Transform applies a transformation function to a string
func Transform(input string, fn func(string) string) string {
	return fn(input)
}

// Reverse returns the reversed string
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ToTitle converts string to title case
func ToTitle(s string) string {
	return strings.Title(strings.ToLower(s))
}

// Greet returns a greeting message
func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}