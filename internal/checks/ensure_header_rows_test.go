package checks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureHeader_Metadata(t *testing.T) {
	c := ensureHeader{}
	if c.Name() != "ensure-header-and-rows" {
		t.Fatalf("Name() = %q", c.Name())
	}
	if !c.FailFast() {
		t.Fatalf("FailFast() should be true")
	}
	if c.Priority() != 3 {
		t.Fatalf("Priority() = %d, want 3", c.Priority())
	}
}

func TestEnsureHeader_Error_CannotOpen(t *testing.T) {
	c := ensureHeader{}
	nonExistent := filepath.Join(t.TempDir(), "nope.csv")
	res := c.Run(nonExistent)
	if res.Status != Error {
		t.Fatalf("Status=%s want ERROR, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_EmptyFile(t *testing.T) {
	c := ensureHeader{}
	path := writeTempFile(t, "")
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "header row is required") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_Header_UsesComma(t *testing.T) {
	c := ensureHeader{}
	path := writeTempFile(t, "term,description,en\na,b,c\n")
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "Header appears to use ','") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_Header_UsesTab(t *testing.T) {
	c := ensureHeader{}
	path := writeTempFile(t, "term\tdescription\ten\nx\ty\tz\n")
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "tab") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_Header_MixedDelimiters(t *testing.T) {
	c := ensureHeader{}
	path := writeTempFile(t, "term;description,en\nx;y,z\n")
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "mixed delimiters") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_Header_MissingRequiredColumns(t *testing.T) {
	c := ensureHeader{}
	path := writeTempFile(t, "term;tags;en\nword;tag;val\n")
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "missing") || !containsLower(res.Message, "description") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_NoDataRows(t *testing.T) {
	c := ensureHeader{}
	path := writeTempFile(t, "term;description;en\n   \n\t\n")
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "Blank lines are not allowed after header. Found at row(s): 2, 3") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_Data_UsesComma(t *testing.T) {
	c := ensureHeader{}
	path := writeTempFile(t, "term;description;en\nword,desc,en\n")
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "CSV parse error on data (check delimiter/quoting)") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_Data_UsesTab(t *testing.T) {
	c := ensureHeader{}
	path := writeTempFile(t, "term;description;en\nword\tdesc\ten\n")
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "CSV parse error on data (check delimiter/quoting)") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_ColumnCountMismatch_OnFirstRow(t *testing.T) {
	c := ensureHeader{}
	// header has 3 cols, first row has 2
	path := writeTempFile(t, "term;description;en\nswitch;device\n")
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "CSV parse error on data (check delimiter/quoting)") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_ColumnCountMismatch_OnSecondRow(t *testing.T) {
	c := ensureHeader{}
	// first data row matches, second mismatches -> should FAIL (we scan multiple rows)
	content := "" +
		"term;description;en\n" +
		"ok;desc;en\n" +
		"bad;only-two-cols\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "CSV parse error on data (check delimiter/quoting): record on line 2") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_RowAfterBlankLines(t *testing.T) {
	c := ensureHeader{}
	content := "" +
		"term;description\n" +
		"\n" +
		"   \n" +
		"valid;desc\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "Blank lines are not allowed after header") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Pass_ValidHeader_AndOneValidRow(t *testing.T) {
	c := ensureHeader{}
	content := "" +
		"term;description;casesensitive;translatable;forbidden;tags;en;en_description;fr;fr_description;de;de_description\n" +
		"switch;Also a device;no;yes;no;network;switch;desc;;;Netwerk switch;\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Pass_QuotedHeader(t *testing.T) {
	c := ensureHeader{}
	content := "\"term\";\"description\"\nfoo;bar\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Pass_SemicolonsInsideQuotes(t *testing.T) {
	c := ensureHeader{}
	content := "term;description\n\"foo;bar\";\"has;semi\"\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Pass_MultilineField(t *testing.T) {
	c := ensureHeader{}
	content := "term;description\nkey;\"line1\nline2\"\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_UnbalancedQuotes(t *testing.T) {
	c := ensureHeader{}
	content := "term;description\n\"oops;bad\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "csv parse error on data (check delimiter/quoting)") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Pass_TabsAndCommasInsideQuotes(t *testing.T) {
	c := ensureHeader{}
	content := "term;description\n\"a,b\tc\";\"x,y\tz\"\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Pass_CRLF(t *testing.T) {
	c := ensureHeader{}
	content := "term;description\r\nalpha;beta\r\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Pass_HeaderWithSpaces(t *testing.T) {
	c := ensureHeader{}
	content := "  term  ;  description \nfoo;bar\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_DataMixedDelimiters_Commas(t *testing.T) {
	c := ensureHeader{}
	content := "term;description;en\nok;desc;en\nbad,desc,en\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "csv parse error on data (check delimiter/quoting)") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_BlankLineInMiddle(t *testing.T) {
	c := ensureHeader{}
	content := "term;description\na;one\n\nb;two\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "blank lines are not allowed after header") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Pass_VeryLongLine(t *testing.T) {
	c := ensureHeader{}
	long := strings.Repeat("x", 200_000)
	content := "term;description\n" + long + ";desc\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_DescriptionTermSwapped(t *testing.T) {
	c := ensureHeader{}
	content := "description;term;tags\ndesc;foo;bar\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "expected first two columns to be 'term;description'") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Fail_TermDescriptionNotAtStart(t *testing.T) {
	c := ensureHeader{}
	content := "tags;term;description\ntag;foo;bar\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Fail || !containsLower(res.Message, "expected first two columns to be 'term;description'") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Pass_TermDescriptionFirst_OthersFollow(t *testing.T) {
	c := ensureHeader{}
	content := "term;description;tags;en;en_description\nfoo;bar;tag1;meaning;desc\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureHeader_Pass_WithBOMOnTerm(t *testing.T) {
	c := ensureHeader{}
	content := "\uFEFFterm;description\nfoo;bar\n"
	path := writeTempFile(t, content)
	res := c.Run(path)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS (BOM stripped), msg=%q", res.Status, res.Message)
	}
}

/*** helpers ***/

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "file.csv")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return p
}

func containsLower(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
