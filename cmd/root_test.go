package cmd

import (
	"testing"
)

func TestRootCmd_HasValidate(t *testing.T) {
	cmd := RootCmd()
	found := false
	for _, c := range cmd.Commands() {
		if c.Name() == "validate" {
			found = true
		}
	}
	if !found {
		t.Fatal("validate subcommand not registered")
	}
}
