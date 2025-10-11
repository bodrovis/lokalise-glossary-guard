package checks

import "testing"

func TestEnsureNoOrphanLangDescriptions_Metadata(t *testing.T) {
	c := ensureNoOrphanLangDescriptions{}
	if c.Name() != "ensure-no-orphan-lang-descriptions" {
		t.Fatalf("Name() = %q", c.Name())
	}
	if c.FailFast() {
		t.Fatalf("FailFast() should be false")
	}
	if c.Priority() != 100 {
		t.Fatalf("Priority() = %d, want 100", c.Priority())
	}
}

func TestEnsureNoOrphanLangDescriptions_Error_EmptyFile(t *testing.T) {
	c := ensureNoOrphanLangDescriptions{}
	res := c.Run([]byte(""), "", nil)
	if res.Status != Error || !containsLower(res.Message, "cannot read header: empty file") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNoOrphanLangDescriptions_Error_MalformedHeader(t *testing.T) {
	c := ensureNoOrphanLangDescriptions{}
	res := c.Run([]byte("onlyone\nx\n"), "", nil)
	if res.Status != Error || !containsLower(res.Message, "malformed header") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNoOrphanLangDescriptions_Pass_NoLangColumns(t *testing.T) {
	c := ensureNoOrphanLangDescriptions{}
	content := "term;description;casesensitive;tags\n" +
		"t;d;yes;a,b\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNoOrphanLangDescriptions_Pass_BaseThenDescription(t *testing.T) {
	c := ensureNoOrphanLangDescriptions{}
	content := "term;description;en;en_description\n" +
		"t;d;hello;desc\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNoOrphanLangDescriptions_Fail_DescriptionWithoutBase(t *testing.T) {
	c := ensureNoOrphanLangDescriptions{}
	content := "term;description;en_description;tags\n" +
		"t;d;desc;x\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail || !containsLower(res.Message, "orphan description column") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
	if !containsLower(res.Message, "en_description") {
		t.Fatalf("Expected to mention 'en_description' in message: %q", res.Message)
	}
}

func TestEnsureNoOrphanLangDescriptions_Fail_DescriptionBeforeBase_TreatedAsOrphan(t *testing.T) {
	// Current implementation is single-pass; description BEFORE base is considered orphan.
	c := ensureNoOrphanLangDescriptions{}
	content := "term;description;fr_description;fr\n" +
		"t;d;desc;bonjour\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail {
		t.Fatalf("Status=%s want FAIL due to description-before-base behavior, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNoOrphanLangDescriptions_Pass_NormalizesLangSeparators(t *testing.T) {
	c := ensureNoOrphanLangDescriptions{}
	// base uses hyphen, description uses underscore (normalize '-' -> '_')
	content := "term;description;pt-BR;pt_BR_description\n" +
		"k;d;oi;desc\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS (separator normalized), msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNoOrphanLangDescriptions_Pass_IgnoresCaseAndSpaces(t *testing.T) {
	c := ensureNoOrphanLangDescriptions{}
	content := "term;description;  EN  ;  en_description  \n" +
		"k;d;hello;desc\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS (case/space-insensitive), msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNoOrphanLangDescriptions_Pass_MultipleLanguages(t *testing.T) {
	c := ensureNoOrphanLangDescriptions{}
	content := "term;description;en;en_description;de_DE;fr;fr_description\n" +
		"t;d;hello;desc;hallo;bonjour;desc-fr\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}
