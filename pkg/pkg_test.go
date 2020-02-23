package pkg

import (
	"fmt"
	"strings"
	"testing"
)

func TestSpeak(t *testing.T) {

	s := []string{"one", "two"}
	fmt.Printf("here: %v\n", Speak(s))
}

func TestWalk(t *testing.T) {
	result := Walk(".")

	if ! strings.Contains(".",result[0]) {
		t.FailNow()
	}
}