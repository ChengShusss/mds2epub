package main

import (
	"os"
	"testing"
)

// Func TestchOsArg proof that os.Args can be changed
func TestChOsArg(t *testing.T) {
	t.Logf("Orgin Args: %+v\n", os.Args)
	os.Args = []string{
		"123", "456",
	}
	t.Logf("Modified Args: %+v\n", os.Args)
}
