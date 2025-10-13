package checks

import (
	"testing"
)

func TestEnsureKnownHeadersWithLangs_Metadata(t *testing.T) {
	c := ensureKnownHeadersWithLangs{}
	if c.Name() != "ensure-known-optional-headers-with-langs" {
		t.Fatalf("Name() = %q", c.Name())
	}
	if c.FailFast() {
		t.Fatalf("FailFast() should be false")
	}
	if c.Priority() != 100 {
		t.Fatalf("Priority() = %d, want 100", c.Priority())
	}
}

func TestEnsureKnownHeadersWithLangs_Error_EmptyFile(t *testing.T) {
	c := ensureKnownHeadersWithLangs{}
	res := c.Run([]byte(""), "", nil)
	if res.Status != Error || !containsLower(res.Message, "cannot read header: empty file") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureKnownHeadersWithLangs_Error_MalformedHeader(t *testing.T) {
	c := ensureKnownHeadersWithLangs{}
	res := c.Run([]byte("onlyone\nx\n"), "", nil)
	if res.Status != Error || !containsLower(res.Message, "malformed header") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureKnownHeadersWithLangs_Pass_NoLangs_OnlyFixedAllowed(t *testing.T) {
	c := ensureKnownHeadersWithLangs{}
	content := "term;description;casesensitive;translatable;forbidden;tags\n" +
		"t;d;yes;no;no;tag1,tag2\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureKnownHeadersWithLangs_Fail_NoLangs_ButLanguageLikeColumns(t *testing.T) {
	c := ensureKnownHeadersWithLangs{}
	content := "term;description;en;fr_description\n" +
		"t;d;hello;bonjour\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail || !containsLower(res.Message, "language-like column(s)") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
	if !containsLower(res.Message, "en") || !containsLower(res.Message, "fr_description") {
		t.Fatalf("Expected to mention 'en' and 'fr_description' in message: %q", res.Message)
	}
}

func TestEnsureKnownHeadersWithLangs_Fail_NoLangs_UnknownColumns(t *testing.T) {
	c := ensureKnownHeadersWithLangs{}
	content := "term;description;weird;another_one\n" +
		"t;d;x;y\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail || !containsLower(res.Message, "unsupported header column(s)") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
	if !containsLower(res.Message, "weird") || !containsLower(res.Message, "another_one") {
		t.Fatalf("Expected to list unknown columns in message: %q", res.Message)
	}
}

func TestEnsureKnownHeadersWithLangs_Pass_WithDeclaredLangs_AllPresent(t *testing.T) {
	c := ensureKnownHeadersWithLangs{}
	langs := []string{"en", "de_DE"}
	content := "term;description;en;en_description;de_DE;tags\n" +
		"k;desc;hello;hello desc;hallo;tag\n"
	res := c.Run([]byte(content), "", langs)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureKnownHeadersWithLangs_Pass_WithDeclaredLangs_DescriptionOptional(t *testing.T) {
	c := ensureKnownHeadersWithLangs{}
	langs := []string{"en"}
	content := "term;description;en\n" +
		"k;desc;hello\n"
	res := c.Run([]byte(content), "", langs)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureKnownHeadersWithLangs_Fail_WithDeclaredLangs_MissingBaseColumn(t *testing.T) {
	c := ensureKnownHeadersWithLangs{}
	langs := []string{"en", "fr"}
	// only fr present, en missing
	content := "term;description;fr;fr_description\n" +
		"k;desc;bonjour;desc\n"
	res := c.Run([]byte(content), "", langs)
	if res.Status != Fail || !containsLower(res.Message, "missing required language column") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
	// should name 'en' as missing
	if !containsLower(res.Message, "en") {
		t.Fatalf("Expected message to include missing language 'en': %q", res.Message)
	}
}

func TestEnsureKnownHeadersWithLangs_Fail_WithDeclaredLangs_UnknownColumns(t *testing.T) {
	c := ensureKnownHeadersWithLangs{}
	langs := []string{"en"}
	content := "term;description;en;unknown_col\n" +
		"k;desc;hello;x\n"
	res := c.Run([]byte(content), "", langs)
	if res.Status != Fail || !containsLower(res.Message, "unsupported header column(s)") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
	if !containsLower(res.Message, "unknown_col") {
		t.Fatalf("Expected to mention unknown_col: %q", res.Message)
	}
	// also expect the message to include declared language list in parentheses
	if !containsLower(res.Message, "(en)") {
		t.Fatalf("Expected declared languages enumerated in message, got: %q", res.Message)
	}
}

func TestEnsureKnownHeadersWithLangs_Pass_NormalizesLangSeparators(t *testing.T) {
	c := ensureKnownHeadersWithLangs{}
	// declare lang with hyphen; header uses underscore
	langs := []string{"de-DE"}
	content := "term;description;de_DE;de_DE_description\n" +
		"k;desc;hallo;beschreibung\n"
	res := c.Run([]byte(content), "", langs)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS (hyphen/underscore normalized), msg=%q", res.Status, res.Message)
	}
}

func TestEnsureKnownHeadersWithLangs_Pass_IgnoresCaseAndBOMInHeaderNames(t *testing.T) {
	c := ensureKnownHeadersWithLangs{}
	langs := []string{"en"}
	// Add BOM before 'term', mixed case for Description, and mixed case fixed columns
	content := "\uFEFFterm;DESCRIPTION;Casesensitive;Translatable;Forbidden;Tags;EN;EN_Description\n" +
		"k;d;yes;no;no;tag;hello;desc\n"
	res := c.Run([]byte(content), "", langs)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}
