package validate

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

func TestNewCmd_CanBeCreatedMultipleTimes(t *testing.T) {
	for i := 0; i < 10; i++ {
		cmd := NewCmd()

		if cmd.Name() != "validate" {
			t.Fatalf("cmd.Name() = %q, want %q", cmd.Name(), "validate")
		}
	}
}

func TestNewCmd_Config(t *testing.T) {
	cmd := NewCmd()

	if cmd.Use != "validate" {
		t.Fatalf("Use = %q, want %q", cmd.Use, "validate")
	}

	if cmd.Short == "" {
		t.Fatal("Short is empty")
	}

	if cmd.Long == "" {
		t.Fatal("Long is empty")
	}

	if cmd.Args == nil {
		t.Fatal("Args = nil, want validator")
	}

	if cmd.PreRunE == nil {
		t.Fatal("PreRunE = nil, want function")
	}

	if cmd.RunE == nil {
		t.Fatal("RunE = nil, want function")
	}
}

func TestNewCmd_Flags(t *testing.T) {
	cmd := NewCmd()

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{"files", "f", "[]"},
		{"langs", "l", "[]"},
		{"parallel", "", ""}, // default depends on runtime.GOMAXPROCS
		{"no-color", "", "false"},
		{"json", "", "false"},
		{"fix", "", "false"},
		{"hard-fail-on-error", "", "false"},
		{"rerun-after-fix", "", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}

			if flag.Shorthand != tt.shorthand {
				t.Fatalf("flag %q shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}

			if tt.defValue != "" && flag.DefValue != tt.defValue {
				t.Fatalf("flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCmd_RejectsPositionalArgs(t *testing.T) {
	cmd := NewCmd()

	err := cmd.Args(cmd, []string{"unexpected"})
	if err == nil {
		t.Fatal("Args() error = nil, want non-nil")
	}
}

func TestValidatePreRun_NoFiles(t *testing.T) {
	opts := defaultCommandOptions()

	err := validatePreRun(&opts)

	if err == nil {
		t.Fatal("error = nil, want no files error")
	}

	want := "no files provided; use --files to specify one or more CSV files"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

func TestValidatePreRun_ExpandsFilesAndPreprocessesLangs(t *testing.T) {
	dir := t.TempDir()

	first := filepath.Join(dir, "a.csv")
	second := filepath.Join(dir, "b.csv")

	mustWriteFileCommandTest(t, first, "a")
	mustWriteFileCommandTest(t, second, "b")

	opts := defaultCommandOptions()
	opts.files = []string{
		first + "," + second,
		first,
	}
	opts.langs = []string{
		"fr,en",
		" en ",
		"lv",
	}

	err := validatePreRun(&opts)
	if err != nil {
		t.Fatalf("validatePreRun() error = %v", err)
	}

	wantFiles := []string{first, second}
	if !slices.Equal(opts.files, wantFiles) {
		t.Fatalf("files = %#v, want %#v", opts.files, wantFiles)
	}

	wantLangs := []string{"en", "fr", "lv"}
	if !slices.Equal(opts.langs, wantLangs) {
		t.Fatalf("langs = %#v, want %#v", opts.langs, wantLangs)
	}
}

func TestValidatePreRun_UsesNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	path := filepath.Join(t.TempDir(), "glossary.csv")
	mustWriteFileCommandTest(t, path, "csv")

	opts := defaultCommandOptions()
	opts.files = []string{path}

	err := validatePreRun(&opts)
	if err != nil {
		t.Fatalf("validatePreRun() error = %v", err)
	}

	if !opts.noColor {
		t.Fatal("opts.noColor = false, want true")
	}
}

func TestValidatePreRun_ExplicitNoColorStaysEnabled(t *testing.T) {
	t.Setenv("NO_COLOR", "")

	path := filepath.Join(t.TempDir(), "glossary.csv")
	mustWriteFileCommandTest(t, path, "csv")

	opts := defaultCommandOptions()
	opts.files = []string{path}
	opts.noColor = true

	err := validatePreRun(&opts)
	if err != nil {
		t.Fatalf("validatePreRun() error = %v", err)
	}

	if !opts.noColor {
		t.Fatal("opts.noColor = false, want true")
	}
}

func TestValidatePreRun_InvalidGlob(t *testing.T) {
	opts := defaultCommandOptions()
	opts.files = []string{"["}

	err := validatePreRun(&opts)

	if err == nil {
		t.Fatal("error = nil, want invalid glob error")
	}
}

func TestValidatePreRun_NoFilesMatched(t *testing.T) {
	opts := defaultCommandOptions()
	opts.files = []string{"", " , "}

	err := validatePreRun(&opts)

	if err == nil {
		t.Fatal("error = nil, want no files matched error")
	}

	want := "no files matched the provided patterns"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

func mustWriteFileCommandTest(t *testing.T, path string, data string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) failed: %v", path, err)
	}
}

func TestNewCmd_ExecuteWithoutFilesReturnsPreRunError(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs(nil)

	err := cmd.Execute()

	if err == nil {
		t.Fatal("error = nil, want no files error")
	}

	want := "no files provided; use --files to specify one or more CSV files"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

func TestNewCmd_RunE(t *testing.T) {
	path := filepath.Join(t.TempDir(), "glossary.csv")

	data := "term;description\nfoo;bar\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write glossary: %v", err)
	}

	cmd := NewCmd()

	var out bytes.Buffer
	var errOut bytes.Buffer

	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	cmd.SetArgs([]string{
		"--files", path,
		"--no-color",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if out.Len() == 0 {
		t.Fatal("stdout is empty, want validation output")
	}

	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestValidatePreRun_NoChecksRegistered(t *testing.T) {
	registered := checks.List()

	t.Cleanup(func() {
		checks.Reset()

		for _, check := range registered {
			if _, err := checks.Register(check); err != nil {
				t.Errorf("restore check %q: %v", check.Name(), err)
			}
		}
	})

	checks.Reset()

	opts := defaultCommandOptions()
	opts.files = []string{
		filepath.Join(t.TempDir(), "glossary.csv"),
	}
	opts.noColor = true

	err := validatePreRun(&opts)
	if err == nil {
		t.Fatal("validatePreRun() error = nil, want non-nil")
	}

	const want = "no checks registered; nothing to run"

	if err.Error() != want {
		t.Fatalf(
			"validatePreRun() error = %q, want %q",
			err.Error(),
			want,
		)
	}
}
