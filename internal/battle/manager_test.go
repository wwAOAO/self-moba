package battle

import (
	"strings"
	"testing"
)

func TestValidatePlayerIDLength(t *testing.T) {
	if err := validatePlayerID(strings.Repeat("\u4f60", MaxPlayerIDLength)); err != nil {
		t.Fatal(err)
	}
	if err := validatePlayerID(strings.Repeat("\u4f60", MaxPlayerIDLength+1)); err == nil {
		t.Fatal("long player id should be rejected")
	}
}
