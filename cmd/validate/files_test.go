package validate

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

func TestHasGlob(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"file.csv", false},
		{"*.csv", true},
		{"file?.csv", true},
		{"[ab].csv", true},
		{"dir/*.csv", true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := hasGlob(tt.in)
			if got != tt.want {
				t.Fatalf("hasGlob(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestWithFixedPostfix(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"glossary.csv", "glossary_fixed.csv"},
		{"glossary_fixed.csv", "glossary_fixed.csv"},
		{"archive.tar.csv", "archive.tar_fixed.csv"},
		{"glossary", "glossary_fixed"},
		{"dir/glossary.csv", "dir/glossary_fixed.csv"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := withFixedPostfix(tt.in)
			if got != tt.want {
				t.Fatalf("withFixedPostfix(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestExpandFiles_DirectFiles(t *testing.T) {
	files, err := expandFiles([]string{
		"one.csv",
		"two.csv",
	})
	if err != nil {
		t.Fatalf("expandFiles returned error: %v", err)
	}

	want := []string{"one.csv", "two.csv"}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("files = %#v, want %#v", files, want)
	}
}

func TestExpandFiles_CommaSeparatedAndTrimmed(t *testing.T) {
	files, err := expandFiles([]string{
		" one.csv, two.csv ",
		"",
		" , three.csv, ",
	})
	if err != nil {
		t.Fatalf("expandFiles returned error: %v", err)
	}

	want := []string{"one.csv", "two.csv", "three.csv"}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("files = %#v, want %#v", files, want)
	}
}

func TestExpandFiles_Deduplicates(t *testing.T) {
	files, err := expandFiles([]string{
		"one.csv,two.csv",
		"one.csv",
		"two.csv",
	})
	if err != nil {
		t.Fatalf("expandFiles returned error: %v", err)
	}

	want := []string{"one.csv", "two.csv"}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("files = %#v, want %#v", files, want)
	}
}

func TestExpandFiles_GlobMatchesFilesOnly(t *testing.T) {
	dir := t.TempDir()

	a := filepath.Join(dir, "a.csv")
	b := filepath.Join(dir, "b.csv")
	subdir := filepath.Join(dir, "sub.csv")

	mustWriteFile(t, a, "a")
	mustWriteFile(t, b, "b")

	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatalf("os.Mkdir failed: %v", err)
	}

	files, err := expandFiles([]string{
		filepath.Join(dir, "*.csv"),
	})
	if err != nil {
		t.Fatalf("expandFiles returned error: %v", err)
	}

	want := []string{a, b}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("files = %#v, want %#v", files, want)
	}
}

func TestExpandFiles_GlobAndDirectDeduplicate(t *testing.T) {
	dir := t.TempDir()

	a := filepath.Join(dir, "a.csv")
	b := filepath.Join(dir, "b.csv")

	mustWriteFile(t, a, "a")
	mustWriteFile(t, b, "b")

	files, err := expandFiles([]string{
		a,
		filepath.Join(dir, "*.csv"),
	})
	if err != nil {
		t.Fatalf("expandFiles returned error: %v", err)
	}

	want := []string{a, b}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("files = %#v, want %#v", files, want)
	}
}

func TestExpandFiles_NoFilesMatched(t *testing.T) {
	_, err := expandFiles([]string{"", " , "})

	if err == nil {
		t.Fatal("error = nil, want error")
	}

	if err.Error() != "no files matched the provided patterns" {
		t.Fatalf("error = %q, want no files matched error", err.Error())
	}
}

func TestExpandFiles_InvalidGlob(t *testing.T) {
	_, err := expandFiles([]string{"["})

	if err == nil {
		t.Fatal("error = nil, want glob syntax error")
	}
}

func TestWriteFixedFileIfNeeded_FixNoneDoesNothing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	finalPath := filepath.Join(dir, "glossary.csv")

	var b strings.Builder
	colors := newColorizer(true)

	err := writeFixedFileIfNeeded(
		&b,
		checks.RunOptions{
			FixMode: checks.FixNone,
		},
		guard.ValidateResponse{
			Fixed:     true,
			FixedData: []byte("fixed"),
			Summary: guard.Summary{
				FinalPath: finalPath,
			},
		},
		colors,
	)
	if err != nil {
		t.Fatalf("writeFixedFileIfNeeded() error = %v", err)
	}

	if b.String() != "" {
		t.Fatalf("output = %q, want empty", b.String())
	}

	assertFileDoesNotExist(
		t,
		filepath.Join(dir, "glossary_fixed.csv"),
	)
}

func TestWriteFixedFileIfNeeded_NotFixedDoesNothing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	finalPath := filepath.Join(dir, "glossary.csv")

	var b strings.Builder
	colors := newColorizer(true)

	err := writeFixedFileIfNeeded(
		&b,
		checks.RunOptions{
			FixMode: checks.FixIfNotPass,
		},
		guard.ValidateResponse{
			Fixed: false,
			Summary: guard.Summary{
				FinalPath: finalPath,
			},
		},
		colors,
	)
	if err != nil {
		t.Fatalf("writeFixedFileIfNeeded() error = %v", err)
	}

	if b.String() != "" {
		t.Fatalf("output = %q, want empty", b.String())
	}

	assertFileDoesNotExist(
		t,
		filepath.Join(dir, "glossary_fixed.csv"),
	)
}

func TestWriteFixedFileIfNeeded_WritesFixedFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	finalPath := filepath.Join(dir, "glossary.csv")
	fixedPath := filepath.Join(dir, "glossary_fixed.csv")

	var b strings.Builder
	colors := newColorizer(true)

	err := writeFixedFileIfNeeded(
		&b,
		checks.RunOptions{
			FixMode: checks.FixIfNotPass,
		},
		guard.ValidateResponse{
			Fixed:     true,
			FixedData: []byte("fixed csv"),
			Summary: guard.Summary{
				FinalPath: finalPath,
			},
		},
		colors,
	)
	if err != nil {
		t.Fatalf("writeFixedFileIfNeeded() error = %v", err)
	}

	got, err := os.ReadFile(fixedPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(got) != "fixed csv" {
		t.Fatalf(
			"fixed file content = %q, want %q",
			got,
			"fixed csv",
		)
	}

	wantOutput := "Info wrote fixed file: " +
		fixedPath +
		" (bytes=9)\n"

	if b.String() != wantOutput {
		t.Fatalf(
			"output = %q, want %q",
			b.String(),
			wantOutput,
		)
	}
}

func TestWriteFixedFileIfNeeded_AlreadyFixedPathDoesNotDoublePostfix(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	finalPath := filepath.Join(dir, "glossary_fixed.csv")

	var b strings.Builder
	colors := newColorizer(true)

	err := writeFixedFileIfNeeded(
		&b,
		checks.RunOptions{
			FixMode: checks.FixIfNotPass,
		},
		guard.ValidateResponse{
			Fixed:     true,
			FixedData: []byte("fixed again"),
			Summary: guard.Summary{
				FinalPath: finalPath,
			},
		},
		colors,
	)
	if err != nil {
		t.Fatalf("writeFixedFileIfNeeded() error = %v", err)
	}

	got, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(got) != "fixed again" {
		t.Fatalf(
			"fixed file content = %q, want %q",
			got,
			"fixed again",
		)
	}

	assertFileDoesNotExist(
		t,
		filepath.Join(dir, "glossary_fixed_fixed.csv"),
	)
}

func TestWriteFixedFileIfNeeded_WriteError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	finalPath := filepath.Join(
		dir,
		"missing",
		"glossary.csv",
	)

	var b strings.Builder
	colors := newColorizer(true)

	err := writeFixedFileIfNeeded(
		&b,
		checks.RunOptions{
			FixMode: checks.FixIfNotPass,
		},
		guard.ValidateResponse{
			Fixed:     true,
			FixedData: []byte("fixed csv"),
			Summary: guard.Summary{
				FinalPath: finalPath,
			},
		},
		colors,
	)

	if err == nil {
		t.Fatal("error = nil, want write error")
	}

	if !strings.Contains(
		b.String(),
		"ERROR writing fixed file:",
	) {
		t.Fatalf(
			"output = %q, want write error message",
			b.String(),
		)
	}
}

func mustWriteFile(t *testing.T, path string, data string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) failed: %v", path, err)
	}
}

func assertFileDoesNotExist(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if err == nil {
		t.Fatalf("file %q exists, want it not to exist", path)
	}

	if !os.IsNotExist(err) {
		t.Fatalf("os.Stat(%q) error = %v, want not exist", path, err)
	}
}

func TestFixedOutputPath_FallbackAndEmpty(t *testing.T) {
	t.Parallel()

	t.Run("falls back to response path", func(t *testing.T) {
		t.Parallel()

		resp := guard.ValidateResponse{
			Path: "foo.csv",
		}

		got, err := fixedOutputPath(resp)
		if err != nil {
			t.Fatalf("fixedOutputPath() error = %v", err)
		}

		if got != "foo_fixed.csv" {
			t.Fatalf("fixedOutputPath() = %q, want %q", got, "foo_fixed.csv")
		}
	})

	t.Run("returns error when all paths are empty", func(t *testing.T) {
		t.Parallel()

		_, err := fixedOutputPath(guard.ValidateResponse{})
		if err == nil {
			t.Fatal("fixedOutputPath() error = nil, want non-nil")
		}

		if err.Error() != "fixed output path is empty" {
			t.Fatalf(
				"fixedOutputPath() error = %q, want %q",
				err.Error(),
				"fixed output path is empty",
			)
		}
	})
}

func TestWriteFixedFileIfNeeded_EmptyOutputPath(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	colors := newColorizer(true)

	resp := guard.ValidateResponse{
		Fixed: true,
	}

	err := writeFixedFileIfNeeded(
		&b,
		checks.RunOptions{
			FixMode: checks.FixAlways,
		},
		resp,
		colors,
	)

	if err == nil {
		t.Fatal("writeFixedFileIfNeeded() error = nil, want non-nil")
	}

	if err.Error() != "fixed output path is empty" {
		t.Fatalf(
			"writeFixedFileIfNeeded() error = %q, want %q",
			err.Error(),
			"fixed output path is empty",
		)
	}

	if !strings.Contains(
		b.String(),
		"ERROR writing fixed file: fixed output path is empty",
	) {
		t.Fatalf(
			"output = %q, want fixed output path error",
			b.String(),
		)
	}
}

func TestExpandFilePattern_StatError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	target := filepath.Join(dir, "missing.csv")
	link := filepath.Join(dir, "broken.csv")

	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	_, err := expandFilePattern(
		filepath.Join(dir, "*.csv"),
	)
	if err == nil {
		t.Fatal("expandFilePattern() error = nil, want stat error")
	}

	if !strings.Contains(err.Error(), `stat matched file`) {
		t.Fatalf(
			"expandFilePattern() error = %q, want stat matched file error",
			err.Error(),
		)
	}

	if !strings.Contains(err.Error(), link) {
		t.Fatalf(
			"expandFilePattern() error = %q, want path %q",
			err.Error(),
			link,
		)
	}
}
