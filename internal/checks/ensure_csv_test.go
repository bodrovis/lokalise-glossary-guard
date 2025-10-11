// ensure_csv_test.go
package checks

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEnsureCSV_Metadata(t *testing.T) {
	c := ensureCSV{}

	if got, want := c.Name(), "ensure-csv-extension"; got != want {
		t.Fatalf("Name() = %q, want %q", got, want)
	}
	if !c.FailFast() {
		t.Fatalf("FailFast() = false, want true")
	}
	if got, want := c.Priority(), 1; got != want {
		t.Fatalf("Priority() = %d, want %d", got, want)
	}
}

func TestEnsureCSV_Run_Pass_csvLower(t *testing.T) {
	c := ensureCSV{}
	fp := filepath.Join("testdata", "glossary.csv")

	res := c.Run(nil, fp, nil)

	if res.Status != Pass {
		t.Fatalf("Status = %s, want %s", res.Status, Pass)
	}
	if !strings.Contains(res.Message, "OK") || !strings.Contains(res.Message, ".csv") {
		t.Fatalf("Message = %q, expected mention of OK and .csv", res.Message)
	}
}

func TestEnsureCSV_Run_Pass_csvUpper(t *testing.T) {
	c := ensureCSV{}
	fp := "ANY/WHERE/GLOSSARY.CSV"

	res := c.Run(nil, fp, nil)

	if res.Status != Pass {
		t.Fatalf("Status = %s, want %s", res.Status, Pass)
	}
}

func TestEnsureCSV_Run_Pass_mixedCaseExt(t *testing.T) {
	c := ensureCSV{}
	fp := "/tmp/mixed.CsV"

	res := c.Run(nil, fp, nil)

	if res.Status != Pass {
		t.Fatalf("Status = %s, want %s", res.Status, Pass)
	}
}

func TestEnsureCSV_Run_Fail_wrongExt(t *testing.T) {
	c := ensureCSV{}
	fp := "/tmp/file.xlsx"

	res := c.Run(nil, fp, nil)

	if res.Status != Fail {
		t.Fatalf("Status = %s, want %s", res.Status, Fail)
	}
	msg := strings.ToLower(res.Message)
	if !strings.Contains(msg, "invalid file extension") {
		t.Fatalf("Message = %q, expected invalid extension note", res.Message)
	}
	if !strings.Contains(res.Message, ".xlsx") {
		t.Fatalf("Message = %q, expected to include .xlsx", res.Message)
	}
}

func TestEnsureCSV_Run_Fail_noExt(t *testing.T) {
	c := ensureCSV{}
	fp := "/path/noext"

	res := c.Run(nil, fp, nil)

	if res.Status != Fail {
		t.Fatalf("Status = %s, want %s", res.Status, Fail)
	}
	if !strings.Contains(res.Message, "(none)") {
		t.Fatalf("Message = %q, expected to include (none) for empty extension", res.Message)
	}
}

func TestEnsureCSV_Run_TrimmedPathSpaces(t *testing.T) {
	c := ensureCSV{}
	fp := "   /tmp/space.csv   "

	res := c.Run(nil, fp, nil)

	if res.Status != Pass {
		t.Fatalf("Status = %s, want %s (should trim spaces)", res.Status, Pass)
	}
}

func TestEnsureCSV_Run_WindowsStylePath(t *testing.T) {
	// Even on non-Windows, Ext should still see ".csv" after the last dot.
	c := ensureCSV{}
	var fp string
	if runtime.GOOS == "windows" {
		fp = `C:\Users\me\Desktop\glossary.csv`
	} else {
		// Simulate a Windows-ish path as a plain string
		fp = `C:\Users\me\Desktop\glossary.csv`
	}

	res := c.Run(nil, fp, nil)

	if res.Status != Pass {
		t.Fatalf("Status = %s, want %s for Windows-like path", res.Status, Pass)
	}
}

func TestEnsureCSV_Run_DotfileCSV(t *testing.T) {
	// Edge: filename starts with a dot but still has .csv extension
	c := ensureCSV{}
	fp := "/tmp/.my.glossary.csv"

	res := c.Run(nil, fp, nil)

	if res.Status != Pass {
		t.Fatalf("Status = %s, want %s", res.Status, Pass)
	}
}

func TestEnsureCSV_Run_EmptyPath(t *testing.T) {
	c := ensureCSV{}
	fp := ""

	res := c.Run(nil, fp, nil)

	if res.Status != Fail {
		t.Fatalf("Status = %s, want %s for empty path", res.Status, Fail)
	}
	if !strings.Contains(res.Message, "(none)") {
		t.Fatalf("Message = %q, expected (none) mention for empty extension", res.Message)
	}
}
