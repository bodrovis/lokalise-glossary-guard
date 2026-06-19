package validate

import (
	"runtime"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

func TestDefaultCommandOptions(t *testing.T) {
	opts := defaultCommandOptions()

	wantMaxParallel := uint(runtime.GOMAXPROCS(0))
	if opts.maxParallel != wantMaxParallel {
		t.Fatalf("maxParallel = %d, want %d", opts.maxParallel, wantMaxParallel)
	}

	if !opts.rerunAfterFix {
		t.Fatal("rerunAfterFix = false, want true")
	}

	if opts.doFix {
		t.Fatal("doFix = true, want false")
	}

	if opts.hardFailOnErr {
		t.Fatal("hardFailOnErr = true, want false")
	}

	if opts.jsonOut {
		t.Fatal("jsonOut = true, want false")
	}

	if opts.noColor {
		t.Fatal("noColor = true, want false")
	}

	if opts.files != nil {
		t.Fatalf("files = %#v, want nil", opts.files)
	}

	if opts.langs != nil {
		t.Fatalf("langs = %#v, want nil", opts.langs)
	}
}

func TestBuildRunOptions_DefaultsToNoFix(t *testing.T) {
	opts := commandOptions{
		rerunAfterFix: true,
	}

	got := buildRunOptions(opts)

	if got.FixMode != checks.FixNone {
		t.Fatalf("FixMode = %v, want %v", got.FixMode, checks.FixNone)
	}

	if !got.RerunAfterFix {
		t.Fatal("RerunAfterFix = false, want true")
	}

	if got.HardFailOnErr {
		t.Fatal("HardFailOnErr = true, want false")
	}
}

func TestBuildRunOptions_WithFix(t *testing.T) {
	opts := commandOptions{
		doFix:         true,
		rerunAfterFix: false,
		hardFailOnErr: true,
	}

	got := buildRunOptions(opts)

	if got.FixMode != checks.FixIfNotPass {
		t.Fatalf("FixMode = %v, want %v", got.FixMode, checks.FixIfNotPass)
	}

	if got.RerunAfterFix {
		t.Fatal("RerunAfterFix = true, want false")
	}

	if !got.HardFailOnErr {
		t.Fatal("HardFailOnErr = false, want true")
	}
}

func TestBuildRunOptions_DoesNotMutateInput(t *testing.T) {
	opts := commandOptions{
		files:         []string{"a.csv"},
		langs:         []string{"en"},
		maxParallel:   4,
		jsonOut:       true,
		noColor:       true,
		doFix:         true,
		hardFailOnErr: true,
		rerunAfterFix: false,
	}

	_ = buildRunOptions(opts)

	if len(opts.files) != 1 || opts.files[0] != "a.csv" {
		t.Fatalf("files mutated: %#v", opts.files)
	}

	if len(opts.langs) != 1 || opts.langs[0] != "en" {
		t.Fatalf("langs mutated: %#v", opts.langs)
	}

	if opts.maxParallel != 4 {
		t.Fatalf("maxParallel mutated: %d", opts.maxParallel)
	}

	if !opts.jsonOut {
		t.Fatal("jsonOut mutated to false")
	}

	if !opts.noColor {
		t.Fatal("noColor mutated to false")
	}

	if !opts.doFix {
		t.Fatal("doFix mutated to false")
	}

	if !opts.hardFailOnErr {
		t.Fatal("hardFailOnErr mutated to false")
	}

	if opts.rerunAfterFix {
		t.Fatal("rerunAfterFix mutated to true")
	}
}
