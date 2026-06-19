package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd_Config(t *testing.T) {
	root := RootCmd()

	if root.Use != "glossary-guard" {
		t.Fatalf("Use = %q, want %q", root.Use, "glossary-guard")
	}

	if root.Short != "Validate Lokalise glossary CSVs" {
		t.Fatalf("Short = %q, want %q", root.Short, "Validate Lokalise glossary CSVs")
	}

	if !root.SilenceUsage {
		t.Fatal("SilenceUsage = false, want true")
	}

	if !root.SilenceErrors {
		t.Fatal("SilenceErrors = false, want true")
	}

	if !root.TraverseChildren {
		t.Fatal("TraverseChildren = false, want true")
	}

	if !strings.Contains(root.Long, "validates CSV files") {
		t.Fatalf("Long = %q, want it to describe CSV validation", root.Long)
	}
}

func TestRootCmd_HasValidateCommand(t *testing.T) {
	root := RootCmd()

	cmd, _, err := root.Find([]string{"validate"})
	if err != nil {
		t.Fatalf("Find validate returned error: %v", err)
	}

	if cmd == nil {
		t.Fatal("validate command not found")
	}

	if cmd.Name() != "validate" {
		t.Fatalf("command name = %q, want %q", cmd.Name(), "validate")
	}
}

func TestRootCmd_HasVersionCommand(t *testing.T) {
	root := RootCmd()

	cmd, _, err := root.Find([]string{"version"})
	if err != nil {
		t.Fatalf("Find version returned error: %v", err)
	}

	if cmd == nil {
		t.Fatal("version command not found")
	}

	if cmd.Name() != "version" {
		t.Fatalf("command name = %q, want %q", cmd.Name(), "version")
	}
}

func TestRootCmd_VersionOutput(t *testing.T) {
	oldVersion := version
	oldCommit := commit
	oldDate := date
	t.Cleanup(func() {
		version = oldVersion
		commit = oldCommit
		date = oldDate
	})

	version = "v1.2.3"
	commit = "abc123"
	date = "2026-06-19"

	root := RootCmd()

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	want := "glossary-guard v1.2.3\ncommit: abc123\nbuilt at: 2026-06-19\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestRootCmd_HasHiddenGendocsCommand(t *testing.T) {
	root := RootCmd()

	cmd, _, err := root.Find([]string{"gendocs"})
	if err != nil {
		t.Fatalf("Find gendocs returned error: %v", err)
	}

	if cmd == nil {
		t.Fatal("gendocs command not found")
	}

	if cmd.Name() != "gendocs" {
		t.Fatalf("command name = %q, want %q", cmd.Name(), "gendocs")
	}

	if !cmd.Hidden {
		t.Fatal("gendocs Hidden = false, want true")
	}
}
