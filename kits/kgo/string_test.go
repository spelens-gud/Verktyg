package kgo

import (
	"testing"
)

func TestOrStrings(t *testing.T) {
	if InitStrings("", "a", "b", "c") != "a" {
		t.Fatalf("")
	}
	if InitStrings() != "" {
		t.Fatalf("")
	}
	if InitStrings("") != "" {
		t.Fatalf("")
	}
	if InitStrings("a") != "a" {
		t.Fatalf("")
	}
}

func TestEnsureStrings(t *testing.T) {
	var s string
	InitStringPtr(&s, "", "", "a")
	if s != "a" {
		t.Fatalf("")
	}
	InitStringPtr(&s, "", "", "b", "c", "d")
	if s != "a" {
		t.Fatalf("")
	}
	s = ""
	InitStringPtr(&s, "", "", "b", "c", "d")
	if s != "b" {
		t.Fatalf("")
	}
}
