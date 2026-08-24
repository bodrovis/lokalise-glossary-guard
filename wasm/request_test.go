//go:build js && wasm

package main

import (
	"slices"
	"strings"
	"syscall/js"
	"testing"
)

func TestValidateRequestFromArgs_MissingInput(t *testing.T) {
	req, err := validateRequestFromArgs(nil)

	if err == nil {
		t.Fatal("error = nil, want error")
	}

	if err.Error() != "missing input" {
		t.Fatalf("error = %q, want %q", err.Error(), "missing input")
	}

	if req.Path != "" {
		t.Fatalf("Path = %q, want empty", req.Path)
	}
}

func TestValidateRequestFromArgs_ValidJSONString(t *testing.T) {
	req, err := validateRequestFromArgs([]js.Value{
		js.ValueOf(`{
			"path": "glossary.csv",
			"data": "term;description;en\nhello;Hello;hello\n",
			"langs": ["fr", "en,lv"],
			"fix": true,
			"rerun_after_fix": false,
			"hard_fail_on_error": true
		}`),
	})
	if err != nil {
		t.Fatalf("validateRequestFromArgs returned error: %v", err)
	}

	if req.Path != "glossary.csv" {
		t.Fatalf("Path = %q, want %q", req.Path, "glossary.csv")
	}

	if req.Text != "term;description;en\nhello;Hello;hello\n" {
		t.Fatalf("Text = %q, want CSV text", req.Text)
	}

	if req.Data != nil {
		t.Fatalf("Data = %#v, want nil", req.Data)
	}

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

	wantLangs := []string{"fr", "en,lv"}
	if !slices.Equal(req.Langs, wantLangs) {
		t.Fatalf("Langs = %#v, want %#v", req.Langs, wantLangs)
	}
}

func TestValidateRequestFromArgs_ValidJSObject(t *testing.T) {
	req, err := validateRequestFromArgs([]js.Value{
		js.ValueOf(map[string]any{
			"path":               "glossary.csv",
			"data":               "term;description;en\nhello;Hello;hello\n",
			"langs":              []any{"fr", "en,lv"},
			"fix":                true,
			"rerun_after_fix":    false,
			"hard_fail_on_error": true,
		}),
	})
	if err != nil {
		t.Fatalf("validateRequestFromArgs returned error: %v", err)
	}

	if req.Path != "glossary.csv" {
		t.Fatalf("Path = %q, want %q", req.Path, "glossary.csv")
	}

	if req.Text == "" {
		t.Fatal("Text is empty, want CSV text")
	}

	if req.Data != nil {
		t.Fatalf("Data = %#v, want nil", req.Data)
	}

	if !req.Fix {
		t.Fatal("Fix = false, want true")
	}

	if req.RerunAfterFix == nil || *req.RerunAfterFix {
		t.Fatalf("RerunAfterFix = %#v, want pointer to false", req.RerunAfterFix)
	}

	if !req.HardFailOnErr {
		t.Fatal("HardFailOnErr = false, want true")
	}

	wantLangs := []string{"fr", "en,lv"}
	if !slices.Equal(req.Langs, wantLangs) {
		t.Fatalf("Langs = %#v, want %#v", req.Langs, wantLangs)
	}
}

func TestValidateRequestFromArgs_InvalidJSONString(t *testing.T) {
	_, err := validateRequestFromArgs([]js.Value{
		js.ValueOf("not json"),
	})

	if err == nil {
		t.Fatal("error = nil, want error")
	}

	if !strings.Contains(err.Error(), "invalid input json") {
		t.Fatalf("error = %q, want invalid input json error", err.Error())
	}

	if !strings.Contains(err.Error(), "pass an object like {data: csvText}") {
		t.Fatalf("error = %q, want usage hint", err.Error())
	}
}

func TestValidateRequestFromArgs_InvalidJSONShape(t *testing.T) {
	_, err := validateRequestFromArgs([]js.Value{
		js.ValueOf(`{"langs": "en"}`),
	})

	if err == nil {
		t.Fatal("error = nil, want error")
	}

	if !strings.Contains(err.Error(), "invalid input json") {
		t.Fatalf("error = %q, want invalid input json error", err.Error())
	}
}

func TestInputToJSON_String(t *testing.T) {
	input := `{"data":"csv text"}`

	got, err := inputToJSON(js.ValueOf(input))
	if err != nil {
		t.Fatalf("inputToJSON returned error: %v", err)
	}

	if got != input {
		t.Fatalf("inputToJSON = %q, want %q", got, input)
	}
}

func TestInputToJSON_Object(t *testing.T) {
	got, err := inputToJSON(js.ValueOf(map[string]any{
		"path": "glossary.csv",
		"data": "csv text",
		"fix":  true,
	}))
	if err != nil {
		t.Fatalf("inputToJSON returned error: %v", err)
	}

	for _, want := range []string{
		`"path":"glossary.csv"`,
		`"data":"csv text"`,
		`"fix":true`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("inputToJSON = %q, want it to contain %q", got, want)
		}
	}
}

func TestInputToJSON_Undefined(t *testing.T) {
	_, err := inputToJSON(js.Undefined())

	if err == nil {
		t.Fatal("error = nil, want error")
	}

	if err.Error() != "input must be a JSON string or object" {
		t.Fatalf("error = %q, want %q", err.Error(), "input must be a JSON string or object")
	}
}

func TestInputToJSON_Null(t *testing.T) {
	_, err := inputToJSON(js.Null())

	if err == nil {
		t.Fatal("error = nil, want error")
	}

	if err.Error() != "input must be a JSON string or object" {
		t.Fatalf("error = %q, want %q", err.Error(), "input must be a JSON string or object")
	}
}

func TestValidateRequestFromArgs_NullJSONString(t *testing.T) {
	_, err := validateRequestFromArgs([]js.Value{
		js.ValueOf("null"),
	})

	if err == nil {
		t.Fatal("error = nil, want error")
	}

	if err.Error() != "input must be a JSON string or object" {
		t.Fatalf("error = %q, want input error", err.Error())
	}
}

func TestInputToJSON_Function(t *testing.T) {
	fn := js.Global().Get("Function").New("return 1")

	_, err := inputToJSON(fn)

	if err == nil {
		t.Fatal("error = nil, want error")
	}

	if err.Error() != "input must be a JSON string or object" {
		t.Fatalf("error = %q, want %q", err.Error(), "input must be a JSON string or object")
	}
}

func TestInputToJSON_CyclicObjectReturnsError(t *testing.T) {
	obj := js.Global().Get("Object").New()
	obj.Set("self", obj)

	got, err := inputToJSON(obj)

	if err == nil {
		t.Fatal("error = nil, want serialization error")
	}

	if got != "" {
		t.Fatalf("inputToJSON() = %q, want empty", got)
	}

	if !strings.Contains(err.Error(), "failed to serialize input") {
		t.Fatalf(
			"error = %q, want failed to serialize input error",
			err.Error(),
		)
	}
}

func TestValidateRequestFromArgs_CyclicObjectReturnsError(t *testing.T) {
	obj := js.Global().Get("Object").New()
	obj.Set("self", obj)

	req, err := validateRequestFromArgs([]js.Value{obj})

	if err == nil {
		t.Fatal("error = nil, want serialization error")
	}

	if !strings.Contains(err.Error(), "failed to serialize input") {
		t.Fatalf(
			"error = %q, want failed to serialize input error",
			err.Error(),
		)
	}

	if req.Path != "" {
		t.Fatalf("Path = %q, want empty", req.Path)
	}
}

func TestValidateGlossaryGuard_CyclicObjectReturnsError(t *testing.T) {
	obj := js.Global().Get("Object").New()
	obj.Set("self", obj)

	raw := validateGlossaryGuard(
		js.Undefined(),
		[]js.Value{obj},
	)

	env := decodeWASMEnvelope(t, raw)

	if env.OK {
		t.Fatal("OK = true, want false")
	}

	if env.Result != nil {
		t.Fatalf("Result = %#v, want nil", env.Result)
	}

	if !strings.Contains(env.Error, "failed to serialize input") {
		t.Fatalf(
			"Error = %q, want serialization error",
			env.Error,
		)
	}
}
