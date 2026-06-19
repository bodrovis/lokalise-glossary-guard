package validate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

func TestWorkerCount(t *testing.T) {
	tests := []struct {
		name        string
		maxParallel uint
		filesCount  int
		wantMin     int
		wantMax     int
	}{
		{
			name:        "uses one worker for empty files",
			maxParallel: 1,
			filesCount:  0,
			wantMin:     1,
			wantMax:     1,
		},
		{
			name:        "caps workers by files count",
			maxParallel: 8,
			filesCount:  3,
			wantMin:     3,
			wantMax:     3,
		},
		{
			name:        "uses requested parallelism",
			maxParallel: 2,
			filesCount:  5,
			wantMin:     2,
			wantMax:     2,
		},
		{
			name:        "zero parallel falls back to gomaxprocs but at least one",
			maxParallel: 0,
			filesCount:  5,
			wantMin:     1,
			wantMax:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := workerCount(tt.maxParallel, tt.filesCount)

			if got < tt.wantMin || got > tt.wantMax {
				t.Fatalf("workerCount(%d, %d) = %d, want between %d and %d",
					tt.maxParallel,
					tt.filesCount,
					got,
					tt.wantMin,
					tt.wantMax,
				)
			}
		})
	}
}

func TestRunOneFile_ReadError(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	missing := filepath.Join(t.TempDir(), "missing.csv")

	oc := runOneFile(context.Background(), 0, missing, nil, "---", checks.RunOptions{
		FixMode:       checks.FixNone,
		RerunAfterFix: true,
	})

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

func TestRunOneFile_ValidFilePasses(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	path := writeTempCSV(t, validRunCSV())

	oc := runOneFile(context.Background(), 0, path, nil, "---", checks.RunOptions{
		FixMode:       checks.FixNone,
		RerunAfterFix: true,
	})

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
			t.Fatalf("Output = %q, want it to contain %q", oc.Output, want)
		}
	}
}

func TestRunOneFile_InvalidFileFailsValidation(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	path := writeTempCSV(t, "term,description,en\nsession,A login session,session\n")

	oc := runOneFile(context.Background(), 0, path, nil, "---", checks.RunOptions{
		FixMode:       checks.FixNone,
		RerunAfterFix: true,
	})

	if oc.Passed != 0 {
		t.Fatalf("Passed = %d, want 0; output:\n%s", oc.Passed, oc.Output)
	}

	if oc.Failed != 1 {
		t.Fatalf("Failed = %d, want 1; output:\n%s", oc.Failed, oc.Output)
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

func TestRunFiles_PreservesInputOrder(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	dir := t.TempDir()

	first := filepath.Join(dir, "first.csv")
	second := filepath.Join(dir, "second.csv")
	third := filepath.Join(dir, "third.csv")

	mustWriteFileRunTest(t, first, validRunCSV())
	mustWriteFileRunTest(t, second, validRunCSV())
	mustWriteFileRunTest(t, third, validRunCSV())

	files := []string{first, second, third}

	outcomes := runFiles(context.Background(), files, nil, "---", 3, checks.RunOptions{
		FixMode:       checks.FixNone,
		RerunAfterFix: true,
	})

	if len(outcomes) != len(files) {
		t.Fatalf("len(outcomes) = %d, want %d", len(outcomes), len(files))
	}

	for i, oc := range outcomes {
		if oc.Idx != i {
			t.Fatalf("outcomes[%d].Idx = %d, want %d", i, oc.Idx, i)
		}

		if oc.Path != files[i] {
			t.Fatalf("outcomes[%d].Path = %q, want %q", i, oc.Path, files[i])
		}

		if oc.Passed != 1 {
			t.Fatalf("outcomes[%d].Passed = %d, want 1; output:\n%s", i, oc.Passed, oc.Output)
		}
	}
}

func TestRunFiles_EmptyInput(t *testing.T) {
	outcomes := runFiles(context.Background(), nil, nil, "---", 4, checks.RunOptions{
		FixMode:       checks.FixNone,
		RerunAfterFix: true,
	})

	if len(outcomes) != 0 {
		t.Fatalf("len(outcomes) = %d, want 0", len(outcomes))
	}
}

func TestRunFiles_PreCanceledContextMarksOutcomesAsOperationalErrors(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	dir := t.TempDir()

	first := filepath.Join(dir, "first.csv")
	second := filepath.Join(dir, "second.csv")

	mustWriteFileRunTest(t, first, validRunCSV())
	mustWriteFileRunTest(t, second, validRunCSV())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	outcomes := runFiles(ctx, []string{first, second}, nil, "---", 2, checks.RunOptions{
		FixMode:       checks.FixNone,
		RerunAfterFix: true,
	})

	if len(outcomes) != 2 {
		t.Fatalf("len(outcomes) = %d, want 2", len(outcomes))
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
			t.Fatalf("outcomes[%d].Output = %q, want context canceled", i, oc.Output)
		}
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
