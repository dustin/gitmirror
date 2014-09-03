package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"io"
	"testing"
)

func TestHMACCompare(t *testing.T) {
	tests := []struct {
		data string
		hash string
		eq   bool
	}{
		{"", "sha1=8a3b873f8dcaebf748c60464fb16878b2953d6df", true},
		{"xxx", "sha1=950408a7db2d17330d8a288417c9d38fd8c6bfef", true},
		{"xxx", "sha1=950408a7db2d17330d8a288417c9d38fd8c6bfee", false},
		{"xxx", "", false},
	}

	for _, test := range tests {
		h := hmac.New(sha1.New, []byte{'h', 'i'})
		io.WriteString(h, test.data)
		if checkHMAC(h, test.hash) != test.eq {
			t.Errorf("On %q, expected %v, got %x", test.data, test.eq, h.Sum(nil))
		}
	}
}
