package validate

import (
	"reflect"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

func TestBuildValidateRequest_NoFix(t *testing.T) {
	data := []byte("term;description;en\nhello;Hello;hello\n")
	langs := []string{"en", "fr"}

	req := buildValidateRequest("glossary.csv", data, langs, checks.RunOptions{
		FixMode:       checks.FixNone,
		RerunAfterFix: true,
		HardFailOnErr: false,
	})

	if req.Path != "glossary.csv" {
		t.Fatalf("Path = %q, want %q", req.Path, "glossary.csv")
	}

	if string(req.Data) != string(data) {
		t.Fatalf("Data = %q, want %q", string(req.Data), string(data))
	}

	if !reflect.DeepEqual(req.Langs, langs) {
		t.Fatalf("Langs = %#v, want %#v", req.Langs, langs)
	}

	if req.Fix {
		t.Fatal("Fix = true, want false")
	}

	if req.RerunAfterFix == nil {
		t.Fatal("RerunAfterFix = nil, want pointer")
	}

	if !*req.RerunAfterFix {
		t.Fatal("RerunAfterFix = false, want true")
	}

	if req.HardFailOnErr {
		t.Fatal("HardFailOnErr = true, want false")
	}
}

func TestBuildValidateRequest_WithFix(t *testing.T) {
	req := buildValidateRequest("glossary.csv", []byte("csv"), []string{"en"}, checks.RunOptions{
		FixMode:       checks.FixIfNotPass,
		RerunAfterFix: false,
		HardFailOnErr: true,
	})

	if !req.Fix {
		t.Fatal("Fix = false, want true")
	}

	if req.RerunAfterFix == nil {
		t.Fatal("RerunAfterFix = nil, want pointer")
	}

	if *req.RerunAfterFix {
		t.Fatal("RerunAfterFix = true, want false")
	}

	if !req.HardFailOnErr {
		t.Fatal("HardFailOnErr = false, want true")
	}
}

func TestBuildValidateRequest_RerunAfterFixPointerIsIndependent(t *testing.T) {
	opts := checks.RunOptions{
		FixMode:       checks.FixIfNotPass,
		RerunAfterFix: true,
	}

	req := buildValidateRequest("glossary.csv", []byte("csv"), nil, opts)

	opts.RerunAfterFix = false

	if req.RerunAfterFix == nil {
		t.Fatal("RerunAfterFix = nil, want pointer")
	}

	if !*req.RerunAfterFix {
		t.Fatal("RerunAfterFix changed after opts mutation, want independent captured value")
	}
}
