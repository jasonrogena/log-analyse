package digest

import (
	"testing"
)

func TestCleanUri(t *testing.T) {
	if CleanUri("a/b") != "a/b" {
		t.Error("Expected a/b")
	}

	if CleanUri("a//b") != "a/b" {
		t.Error("Expected a/b")
	}

	if CleanUri("a///b") != "a/b" {
		t.Error("Expected a/b")
	}

	if CleanUri("a/") != "a/" {
		t.Error("Expected a/")
	}
}
