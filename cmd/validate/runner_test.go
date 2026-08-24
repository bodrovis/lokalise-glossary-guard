package validate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

func testFileRunConfig() fileRunConfig {
	return fileRunConfig{
		Separator:   "---",
		MaxParallel: 1,
		Colors:      newColorizer(true),
		Options: checks.RunOptions{
			FixMode:       checks.FixNone,
			RerunAfterFix: true,
		},
	}
}

func TestWorkerCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		maxParallel uint
		filesCount  int
		want        int
	}{
		{
			name:        "uses one worker for empty files",
			maxParallel: 1,
			filesCount:  0,
			want:        1,
		},
		{
			name:        "caps workers by files count",
			maxParallel: 8,
			filesCount:  3,
			want:        3,
		},
		{
			name:        "uses requested parallelism",
			maxParallel: 2,
			filesCount:  5,
			want:        2,
		},
		{
			name:        "zero parallel uses gomaxprocs",
			maxParallel: 0,
			filesCount:  5,
			want:        min(runtime.GOMAXPROCS(0), 5),
		},
		{
			name:        "large parallelism does not overflow",
			maxParallel: ^uint(0),
			filesCount:  3,
			want:        3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := workerCount(tt.maxParallel, tt.filesCount)

			if got != tt.want {
				t.Fatalf(
					"workerCount(%d, %d) = %d, want %d",
					tt.maxParallel,
					tt.filesCount,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestFileRunConfig_RunOneFileReadError(t *testing.T) {
	t.Parallel()

	cfg := testFileRunConfig()
	missing := filepath.Join(t.TempDir(), "missing.csv")

	oc := cfg.runOneFile(
		context.Background(),
		0,
		missing,
	)

	if oc.Idx != 0 {
		t.Fatalf("Idx = %d, want 0", oc.Idx)
	}
	if oc.Path != missing {
		t.Fatalf("Path = %q, want %q", oc.Path, missing)
	}
	if oc.Errored != 1 {
		t.Fatalf("Errored = %d, want 1", oc.Errored)
	}
	if !oc.HadOpErr {
		t.Fatal("HadOpErr = false, want true")
	}
	if oc.HadValFail {
		t.Fatal("HadValFail = true, want false")
	}
	if !strings.Contains(oc.Output, "ERROR:") {
		t.Fatalf("Output = %q, want ERROR", oc.Output)
	}
}

func TestFileRunConfig_RunOneFileValidFilePasses(t *testing.T) {
	t.Parallel()

	cfg := testFileRunConfig()
	path := writeTempCSV(t, validRunCSV())

	oc := cfg.runOneFile(
		context.Background(),
		0,
		path,
	)

	if oc.Path != path {
		t.Fatalf("Path = %q, want %q", oc.Path, path)
	}
	if oc.Passed != 1 {
		t.Fatalf("Passed = %d, want 1; output:\n%s", oc.Passed, oc.Output)
	}
	if oc.Failed != 0 {
		t.Fatalf("Failed = %d, want 0; output:\n%s", oc.Failed, oc.Output)
	}
	if oc.Errored != 0 {
		t.Fatalf("Errored = %d, want 0; output:\n%s", oc.Errored, oc.Output)
	}
	if oc.HadOpErr {
		t.Fatal("HadOpErr = true, want false")
	}
	if oc.HadValFail {
		t.Fatal("HadValFail = true, want false")
	}
	if oc.Summary == nil {
		t.Fatal("Summary = nil, want summary")
	}

	for _, want := range []string{
		"Validating:",
		"Summary for ",
		"Result: PASSED",
	} {
		if !strings.Contains(oc.Output, want) {
			t.Fatalf(
				"Output = %q, want it to contain %q",
				oc.Output,
				want,
			)
		}
	}
}

func TestFileRunConfig_RunOneFileInvalidFileFailsValidation(t *testing.T) {
	t.Parallel()

	cfg := testFileRunConfig()
	path := writeTempCSV(
		t,
		"term,description,en\nsession,A login session,session\n",
	)

	oc := cfg.runOneFile(
		context.Background(),
		0,
		path,
	)

	if oc.Passed != 0 {
		t.Fatalf("Passed = %d, want 0", oc.Passed)
	}
	if oc.Failed != 1 {
		t.Fatalf("Failed = %d, want 1", oc.Failed)
	}
	if !oc.HadValFail {
		t.Fatal("HadValFail = false, want true")
	}
	if oc.HadOpErr {
		t.Fatal("HadOpErr = true, want false")
	}
	if oc.Summary == nil {
		t.Fatal("Summary = nil, want summary")
	}
	if !strings.Contains(oc.Output, "Result: FAILED") {
		t.Fatalf("Output = %q, want FAILED result", oc.Output)
	}
}

func TestFileRunConfig_RunFilesPreservesInputOrder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	files := []string{
		filepath.Join(dir, "first.csv"),
		filepath.Join(dir, "second.csv"),
		filepath.Join(dir, "third.csv"),
	}

	for _, path := range files {
		mustWriteFileRunTest(t, path, validRunCSV())
	}

	cfg := testFileRunConfig()
	cfg.MaxParallel = 3

	outcomes := cfg.runFiles(
		context.Background(),
		files,
	)

	if len(outcomes) != len(files) {
		t.Fatalf(
			"len(outcomes) = %d, want %d",
			len(outcomes),
			len(files),
		)
	}

	for i, oc := range outcomes {
		if oc.Idx != i {
			t.Fatalf(
				"outcomes[%d].Idx = %d, want %d",
				i,
				oc.Idx,
				i,
			)
		}
		if oc.Path != files[i] {
			t.Fatalf(
				"outcomes[%d].Path = %q, want %q",
				i,
				oc.Path,
				files[i],
			)
		}
		if oc.Passed != 1 {
			t.Fatalf(
				"outcomes[%d].Passed = %d, want 1",
				i,
				oc.Passed,
			)
		}
	}
}

func TestFileRunConfig_RunFilesEmptyInput(t *testing.T) {
	t.Parallel()

	cfg := testFileRunConfig()
	cfg.MaxParallel = 4

	outcomes := cfg.runFiles(
		context.Background(),
		nil,
	)

	if len(outcomes) != 0 {
		t.Fatalf("len(outcomes) = %d, want 0", len(outcomes))
	}
}

func TestFileRunConfig_RunFilesPreCanceledContext(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	files := []string{
		filepath.Join(dir, "first.csv"),
		filepath.Join(dir, "second.csv"),
	}

	for _, path := range files {
		mustWriteFileRunTest(t, path, validRunCSV())
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := testFileRunConfig()
	cfg.MaxParallel = 2

	outcomes := cfg.runFiles(ctx, files)

	if len(outcomes) != len(files) {
		t.Fatalf(
			"len(outcomes) = %d, want %d",
			len(outcomes),
			len(files),
		)
	}

	for i, oc := range outcomes {
		if oc.Idx != i {
			t.Fatalf("outcomes[%d].Idx = %d, want %d", i, oc.Idx, i)
		}
		if oc.Errored != 1 {
			t.Fatalf("outcomes[%d].Errored = %d, want 1", i, oc.Errored)
		}
		if !oc.HadOpErr {
			t.Fatalf("outcomes[%d].HadOpErr = false, want true", i)
		}
		if !strings.Contains(oc.Output, "context canceled") {
			t.Fatalf(
				"outcomes[%d].Output = %q, want context canceled",
				i,
				oc.Output,
			)
		}
	}
}

func TestRunValidate_ValidFile(t *testing.T) {
	t.Parallel()

	path := writeTempCSV(t, validRunCSV())

	var out strings.Builder
	var errOut strings.Builder

	err := runValidate(
		context.Background(),
		validateRunConfig{
			Files:       []string{path},
			MaxParallel: 1,
			NoColor:     true,
			Out:         &out,
			ErrOut:      &errOut,
			Options: checks.RunOptions{
				FixMode:       checks.FixNone,
				RerunAfterFix: true,
			},
		},
	)
	if err != nil {
		t.Fatalf("runValidate() error = %v", err)
	}

	if out.Len() == 0 {
		t.Fatal("output is empty, want validation output")
	}

	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestFileRunConfig_ValidateFileCanceledContext(t *testing.T) {
	t.Parallel()

	path := writeTempCSV(t, validRunCSV())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := testFileRunConfig()

	resp, err := cfg.validateFile(ctx, path)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"validateFile() error = %v, want %v",
			err,
			context.Canceled,
		)
	}

	if !reflect.ValueOf(resp).IsZero() {
		t.Fatalf(
			"validateFile() response = %#v, want zero value",
			resp,
		)
	}
}

func TestFileRunConfig_RunOneFileFixedFileWriteError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "glossary.csv")

	data := "term;description;en;fr\n\nsession;A login session;session;session\n"

	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	// Block the expected output path with a directory so os.WriteFile
	// fails deterministically.
	fixedPath := filepath.Join(dir, "glossary_fixed.csv")
	if err := os.Mkdir(fixedPath, 0o755); err != nil {
		t.Fatalf("create blocking directory: %v", err)
	}

	cfg := testFileRunConfig()
	cfg.Options.FixMode = checks.FixAlways

	oc := cfg.runOneFile(
		context.Background(),
		0,
		path,
	)

	if !oc.HadOpErr {
		t.Fatal("HadOpErr = false, want true")
	}

	if oc.Errored == 0 {
		t.Fatal("Errored = 0, want at least 1")
	}

	if !strings.Contains(oc.Output, "writing fixed file") {
		t.Fatalf(
			"Output = %q, want fixed file write error",
			oc.Output,
		)
	}
}

func writeTempCSV(t *testing.T, data string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "glossary.csv")
	mustWriteFileRunTest(t, path, data)

	return path
}

func mustWriteFileRunTest(t *testing.T, path string, data string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) failed: %v", path, err)
	}
}

func validRunCSV() string {
	return "term;description;en;fr\nsession;A login session;session;session\n"
}
