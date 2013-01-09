package main

import (
	"testing"
)

const sample = `<https://api.github.com/orgs/couchbaselabs/repos?page=2>; rel="next", <https://api.github.com/orgs/couchbaselabs/repos?page=5>; rel="last"`

func TestLinkParsing(t *testing.T) {
	l := parseLink(sample)

	exp := map[string]string{
		"next": "https://api.github.com/orgs/couchbaselabs/repos?page=2",
		"last": "https://api.github.com/orgs/couchbaselabs/repos?page=5",
	}

	if len(exp) != len(l) {
		t.Errorf("Expected %v items, got %v", len(exp), len(l))
	}

	for k, v := range exp {
		if l[k] != v {
			t.Errorf("Expected %v for %v, got %v", v, k, l[k])
		}
	}
}

func TestEmptyLinkParsing(t *testing.T) {
	l := parseLink("")
	if len(l) != 0 {
		t.Errorf("Parsed nothing into %v", l)
	}
}
