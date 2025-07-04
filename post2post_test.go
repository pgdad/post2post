package post2post

import (
	"strings"
	"testing"
)

func TestTransform(t *testing.T) {
	input := "hello"
	expected := "HELLO"
	result := Transform(input, strings.ToUpper)
	
	if result != expected {
		t.Errorf("Transform() = %v, want %v", result, expected)
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "olleh"},
		{"world", "dlrow"},
		{"", ""},
		{"a", "a"},
		{"🌟", "🌟"},
	}
	
	for _, test := range tests {
		result := Reverse(test.input)
		if result != test.expected {
			t.Errorf("Reverse(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestToTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "Hello World"},
		{"HELLO WORLD", "Hello World"},
		{"", ""},
		{"a", "A"},
	}
	
	for _, test := range tests {
		result := ToTitle(test.input)
		if result != test.expected {
			t.Errorf("ToTitle(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestGreet(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"Alice", "Hello, Alice!"},
		{"Bob", "Hello, Bob!"},
		{"", "Hello, !"},
	}
	
	for _, test := range tests {
		result := Greet(test.name)
		if result != test.expected {
			t.Errorf("Greet(%q) = %q, want %q", test.name, result, test.expected)
		}
	}
}