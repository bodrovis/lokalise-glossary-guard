package checks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureUTF8_Metadata(t *testing.T) {
	c := ensureUTF8{}
	if got, want := c.Name(), "ensure-utf8-encoding"; got != want {
		t.Fatalf("Name() = %q, want %q", got, want)
	}
	if !c.FailFast() {
		t.Fatalf("FailFast() = false, want true")
	}
	if got, want := c.Priority(), 2; got != want {
		t.Fatalf("Priority() = %d, want %d", got, want)
	}
}

func TestEnsureUTF8_Run_Pass_SimpleASCII(t *testing.T) {
	c := ensureUTF8{}
	path := writeTemp(t, "hello, world\n")
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status = %s, want %s; msg=%q", res.Status, Pass, res.Message)
	}
}

func TestEnsureUTF8_Run_Pass_UTF8BOM(t *testing.T) {
	c := ensureUTF8{}

	content := string([]byte{0xEF, 0xBB, 0xBF}) + "with bom"
	path := writeTemp(t, content)
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status = %s, want %s; msg=%q", res.Status, Pass, res.Message)
	}
}

func TestEnsureUTF8_Run_Pass_Multibyte(t *testing.T) {
	c := ensureUTF8{}
	path := writeTemp(t, "Привет, 你好!")
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status = %s, want %s; msg=%q", res.Status, Pass, res.Message)
	}
}

func TestEnsureUTF8_Run_Fail_InvalidBytes(t *testing.T) {
	c := ensureUTF8{}

	bad := []byte("ok ")          // valid
	bad = append(bad, 0xFF, 0xFE) // invalid
	bad = append(bad, ' ')        // separator
	bad = append(bad, 0xC3, 0x28) // invalid continuation
	path := writeTempBytes(t, bad)

	res := c.Run(path)
	if res.Status != Fail {
		t.Fatalf("Status = %s, want %s; msg=%q", res.Status, Fail, res.Message)
	}
}

func TestEnsureUTF8_Run_Error_CannotRead(t *testing.T) {
	c := ensureUTF8{}
	nonExistent := filepath.Join(t.TempDir(), "nope.csv")
	res := c.Run(nonExistent)
	if res.Status != Error {
		t.Fatalf("Status = %s, want %s; msg=%q", res.Status, Error, res.Message)
	}
	if res.Message == "" {
		t.Fatalf("expected error message to be non-empty")
	}
}

// helpers

func writeTemp(t *testing.T, s string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "file.csv")
	if err := os.WriteFile(p, []byte(s), 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return p
}

func writeTempBytes(t *testing.T, b []byte) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "file.csv")
	if err := os.WriteFile(p, b, 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return p
}
