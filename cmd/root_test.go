package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCmd_Config(t *testing.T) {
	root := RootCmd()

	if root.Use != "glossary-guard" {
		t.Fatalf("Use = %q, want %q", root.Use, "glossary-guard")
	}

	if root.Short != "Validate Lokalise glossary CSVs" {
		t.Fatalf(
			"Short = %q, want %q",
			root.Short,
			"Validate Lokalise glossary CSVs",
		)
	}

	if root.Long == "" {
		t.Fatal("Long is empty")
	}

	if !root.SilenceUsage {
		t.Fatal("SilenceUsage = false, want true")
	}

	if !root.SilenceErrors {
		t.Fatal("SilenceErrors = false, want true")
	}

	if root.Args == nil {
		t.Fatal("Args = nil, want validator")
	}

	if root.RunE == nil {
		t.Fatal("RunE = nil, want function")
	}
}

func TestRootCmd_HasCommands(t *testing.T) {
	root := RootCmd()

	tests := []struct {
		name   string
		hidden bool
	}{
		{name: "validate"},
		{name: "version"},
		{name: "gendocs", hidden: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, err := root.Find([]string{tt.name})
			if err != nil {
				t.Fatalf("Find(%q) error = %v", tt.name, err)
			}

			if cmd.Name() != tt.name {
				t.Fatalf(
					"command name = %q, want %q",
					cmd.Name(),
					tt.name,
				)
			}

			if cmd.Hidden != tt.hidden {
				t.Fatalf(
					"Hidden = %v, want %v",
					cmd.Hidden,
					tt.hidden,
				)
			}
		})
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

func TestRootCmd_RunEShowsHelp(t *testing.T) {
	t.Parallel()

	cmd := RootCmd()

	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE() error = %v", err)
	}

	if !strings.Contains(out.String(), "glossary-guard") {
		t.Fatalf("help output = %q, want command help", out.String())
	}
}

func TestRootCmd_Gendocs(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	root := RootCmd()

	cmd, _, err := root.Find([]string{"gendocs"})
	if err != nil {
		t.Fatalf("Find(gendocs) error = %v", err)
	}

	if cmd.RunE == nil {
		t.Fatal("gendocs RunE = nil")
	}

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("gendocs RunE() error = %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, "docs"))
	if err != nil {
		t.Fatalf("docs directory: %v", err)
	}

	if !info.IsDir() {
		t.Fatal("docs path is not a directory")
	}

	matches, err := filepath.Glob(filepath.Join(dir, "docs", "*.md"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	if len(matches) == 0 {
		t.Fatal("gendocs generated no Markdown files")
	}
}

func TestRootCmd_RejectsPositionalArgs(t *testing.T) {
	root := RootCmd()

	err := root.Args(root, []string{"unexpected"})
	if err == nil {
		t.Fatal("Args() error = nil, want non-nil")
	}
}

func TestRootCmd_SubcommandsRejectPositionalArgs(t *testing.T) {
	root := RootCmd()

	for _, name := range []string{"version", "gendocs"} {
		t.Run(name, func(t *testing.T) {
			cmd, _, err := root.Find([]string{name})
			if err != nil {
				t.Fatalf("Find(%q) error = %v", name, err)
			}

			if cmd.Args == nil {
				t.Fatalf("%s Args = nil", name)
			}

			if err := cmd.Args(cmd, []string{"unexpected"}); err == nil {
				t.Fatalf("%s Args() error = nil, want non-nil", name)
			}
		})
	}
}
