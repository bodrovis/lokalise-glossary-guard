package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGenerateDocs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "docs")

	root := &cobra.Command{
		Use:   "testcmd",
		Short: "Test command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	root.AddCommand(&cobra.Command{
		Use:   "child",
		Short: "Child command",
		Run:   func(cmd *cobra.Command, args []string) {},
	})

	err := generateDocs(root, dir)
	if err != nil {
		t.Fatalf("generateDocs returned error: %v", err)
	}

	rootDoc := filepath.Join(dir, "testcmd.md")
	childDoc := filepath.Join(dir, "testcmd_child.md")

	assertFileExists(t, rootDoc)
	assertFileExists(t, childDoc)

	rootData, err := os.ReadFile(rootDoc)
	if err != nil {
		t.Fatalf("os.ReadFile root doc failed: %v", err)
	}

	if !strings.Contains(string(rootData), "Test command") {
		t.Fatalf("root doc = %q, want it to contain short description", string(rootData))
	}

	childData, err := os.ReadFile(childDoc)
	if err != nil {
		t.Fatalf("os.ReadFile child doc failed: %v", err)
	}

	if !strings.Contains(string(childData), "Child command") {
		t.Fatalf("child doc = %q, want it to contain short description", string(childData))
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %q to exist: %v", path, err)
	}
}

func TestGenerateDocs_ReturnsMkdirError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "not-a-dir")

	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("os.WriteFile failed: %v", err)
	}

	err := generateDocs(&cobra.Command{Use: "testcmd"}, filepath.Join(file, "docs"))
	if err == nil {
		t.Fatal("error = nil, want mkdir error")
	}
}
