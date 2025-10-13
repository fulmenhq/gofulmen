package main

import (
	"strings"
	"testing"
)

func TestToUpper(t *testing.T) {
	result := strings.ToUpper("hello")
	expected := "HELLO"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
