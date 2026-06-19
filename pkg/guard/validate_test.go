package guard_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

func TestValidateBytes_ValidDataPasses(t *testing.T) {
	resp, err := guard.ValidateBytes(context.Background(), guard.ValidateRequest{
		Path: "glossary.csv",
		Data: []byte(validCSV()),
	})
	if err != nil {
		t.Fatalf("ValidateBytes returned error: %v", err)
	}

	assertStatus(t, resp, guard.StatusPassed)
	assertFlags(t, resp, true, false, false, false)

	if resp.Path != "glossary.csv" {
		t.Fatalf("Path = %q, want %q", resp.Path, "glossary.csv")
	}

	if resp.Fixed {
		t.Fatal("Fixed = true, want false")
	}

	if resp.FixedText != "" {
		t.Fatalf("FixedText = %q, want empty", resp.FixedText)
	}

	if resp.FixedData != nil {
		t.Fatalf("FixedData = %#v, want nil", resp.FixedData)
	}

	if len(resp.Summary.Outcomes) == 0 {
		t.Fatal("Summary.Outcomes is empty, want validation outcomes")
	}
}

func TestValidateBytes_UsesTextWhenDataIsNil(t *testing.T) {
	resp, err := guard.ValidateBytes(context.Background(), guard.ValidateRequest{
		Path: "glossary.csv",
		Text: validCSV(),
	})
	if err != nil {
		t.Fatalf("ValidateBytes returned error: %v", err)
	}

	assertStatus(t, resp, guard.StatusPassed)
	assertFlags(t, resp, true, false, false, false)
}

func TestValidateBytes_PrefersDataOverText(t *testing.T) {
	resp, err := guard.ValidateBytes(context.Background(), guard.ValidateRequest{
		Path: "glossary.csv",

		// Valid data should win.
		Data: []byte(validCSV()),

		// Invalid text should be ignored because Data is present.
		Text: "",
	})
	if err != nil {
		t.Fatalf("ValidateBytes returned error: %v", err)
	}

	assertStatus(t, resp, guard.StatusPassed)
	assertFlags(t, resp, true, false, false, false)
}

func TestValidateBytes_InvalidCSVFailsWithoutGoError(t *testing.T) {
	resp, err := guard.ValidateBytes(context.Background(), guard.ValidateRequest{
		Path: "glossary.csv",
		Data: []byte("term,description\nhello,world\n"),
	})
	if err != nil {
		t.Fatalf("ValidateBytes returned error: %v", err)
	}

	assertStatus(t, resp, guard.StatusFailed)
	assertFlags(t, resp, false, false, true, false)

	if resp.Error != "" {
		t.Fatalf("Error = %q, want empty validation error", resp.Error)
	}

	if resp.Summary.Fail == 0 {
		t.Fatalf("Summary.Fail = %d, want > 0", resp.Summary.Fail)
	}

	if len(resp.Summary.Outcomes) == 0 {
		t.Fatal("Summary.Outcomes is empty, want validation outcomes")
	}
}

func TestValidateBytes_EmptyDataFails(t *testing.T) {
	resp, err := guard.ValidateBytes(context.Background(), guard.ValidateRequest{
		Path: "glossary.csv",
		Data: []byte(""),
	})
	if err != nil {
		t.Fatalf("ValidateBytes returned error: %v", err)
	}

	assertStatus(t, resp, guard.StatusFailed)
	assertFlags(t, resp, false, false, true, false)

	if resp.Summary.Fail == 0 {
		t.Fatalf("Summary.Fail = %d, want > 0", resp.Summary.Fail)
	}
}

func TestValidateBytes_InvalidUTF8Fails(t *testing.T) {
	resp, err := guard.ValidateBytes(context.Background(), guard.ValidateRequest{
		Path: "glossary.csv",
		Data: []byte{0xff, 0xfe, 0xfd},
	})
	if err != nil {
		t.Fatalf("ValidateBytes returned error: %v", err)
	}

	assertStatus(t, resp, guard.StatusFailed)
	assertFlags(t, resp, false, false, true, false)

	if resp.Summary.Fail == 0 {
		t.Fatalf("Summary.Fail = %d, want > 0", resp.Summary.Fail)
	}
}

func TestValidateBytes_InvalidExtensionFails(t *testing.T) {
	resp, err := guard.ValidateBytes(context.Background(), guard.ValidateRequest{
		Path: "glossary.txt",
		Data: []byte(validCSV()),
	})
	if err != nil {
		t.Fatalf("ValidateBytes returned error: %v", err)
	}

	assertStatus(t, resp, guard.StatusFailed)
	assertFlags(t, resp, false, false, true, false)

	if resp.Summary.Fail == 0 {
		t.Fatalf("Summary.Fail = %d, want > 0", resp.Summary.Fail)
	}
}

func TestValidateBytes_FixReturnsFixedDataAndText(t *testing.T) {
	resp, err := guard.ValidateBytes(context.Background(), guard.ValidateRequest{
		Path: "glossary.csv",
		Data: []byte(csvWithEmptyLine()),
		Fix:  true,
	})
	if err != nil {
		t.Fatalf("ValidateBytes returned error: %v", err)
	}

	if !resp.Fixed {
		t.Fatal("Fixed = false, want true")
	}

	if !resp.Summary.AppliedFixes {
		t.Fatal("Summary.AppliedFixes = false, want true")
	}

	if len(resp.FixedData) == 0 {
		t.Fatal("FixedData is empty, want fixed bytes")
	}

	if resp.FixedText == "" {
		t.Fatal("FixedText is empty, want fixed text")
	}

	if resp.FixedText != string(resp.FixedData) {
		t.Fatalf("FixedText != string(FixedData): %q != %q", resp.FixedText, string(resp.FixedData))
	}

	if bytes.Contains(resp.FixedData, []byte("\n\n")) {
		t.Fatalf("FixedData still contains empty line: %q", string(resp.FixedData))
	}
}

func TestValidateBytes_FixDisabledDoesNotReturnFixedData(t *testing.T) {
	resp, err := guard.ValidateBytes(context.Background(), guard.ValidateRequest{
		Path: "glossary.csv",
		Data: []byte(csvWithEmptyLine()),
		Fix:  false,
	})
	if err != nil {
		t.Fatalf("ValidateBytes returned error: %v", err)
	}

	if resp.Fixed {
		t.Fatal("Fixed = true, want false")
	}

	if resp.FixedText != "" {
		t.Fatalf("FixedText = %q, want empty", resp.FixedText)
	}

	if resp.FixedData != nil {
		t.Fatalf("FixedData = %#v, want nil", resp.FixedData)
	}
}

func TestValidateBytes_FixWithExplicitRerunAfterFixFalse(t *testing.T) {
	rerunAfterFix := false

	resp, err := guard.ValidateBytes(context.Background(), guard.ValidateRequest{
		Path:          "glossary.csv",
		Data:          []byte(csvWithEmptyLine()),
		Fix:           true,
		RerunAfterFix: &rerunAfterFix,
	})
	if err != nil {
		t.Fatalf("ValidateBytes returned error: %v", err)
	}

	if !resp.Fixed {
		t.Fatal("Fixed = false, want true")
	}

	if !resp.Summary.AppliedFixes {
		t.Fatal("Summary.AppliedFixes = false, want true")
	}

	if resp.FixedText == "" {
		t.Fatal("FixedText is empty, want fixed text")
	}
}

func TestValidateBytesJSON_ValidResponse(t *testing.T) {
	raw, err := guard.ValidateBytesJSON(context.Background(), guard.ValidateRequest{
		Path: "glossary.csv",
		Data: []byte(validCSV()),
	})
	if err != nil {
		t.Fatalf("ValidateBytesJSON returned error: %v", err)
	}

	var resp guard.ValidateResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("json.Unmarshal failed: %v\nraw: %s", err, raw)
	}

	assertStatus(t, resp, guard.StatusPassed)
	assertFlags(t, resp, true, false, false, false)
}

func TestValidateBytesJSON_FixedDataIsNotSerialized(t *testing.T) {
	raw, err := guard.ValidateBytesJSON(context.Background(), guard.ValidateRequest{
		Path: "glossary.csv",
		Data: []byte(csvWithEmptyLine()),
		Fix:  true,
	})
	if err != nil {
		t.Fatalf("ValidateBytesJSON returned error: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("json.Unmarshal failed: %v\nraw: %s", err, raw)
	}

	if _, ok := body["fixed_data"]; ok {
		t.Fatalf("JSON contains fixed_data, want it omitted: %s", raw)
	}

	fixed, ok := body["fixed"].(bool)
	if !ok || !fixed {
		t.Fatalf("fixed = %#v, want true", body["fixed"])
	}

	fixedText, ok := body["fixed_text"].(string)
	if !ok || fixedText == "" {
		t.Fatalf("fixed_text = %#v, want non-empty string", body["fixed_text"])
	}
}

func TestValidateBytesJSON_ContextCanceledReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	raw, err := guard.ValidateBytesJSON(ctx, guard.ValidateRequest{
		Path: "glossary.csv",
		Data: []byte(validCSV()),
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}

	if raw != nil {
		t.Fatalf("raw = %s, want nil", raw)
	}
}

func validCSV() string {
	return "term;description;en;fr\nsession;A login session;session;session\n"
}

func csvWithEmptyLine() string {
	return "term;description;en;fr\n\nsession;A login session;session;session\n"
}

func assertStatus(t *testing.T, resp guard.ValidateResponse, want guard.ValidateStatus) {
	t.Helper()

	if resp.Status != want {
		t.Fatalf("Status = %q, want %q; response: %#v", resp.Status, want, resp)
	}
}

func assertFlags(t *testing.T, resp guard.ValidateResponse, passed, warned, failed, errored bool) {
	t.Helper()

	if resp.Passed != passed {
		t.Fatalf("Passed = %v, want %v; response: %#v", resp.Passed, passed, resp)
	}

	if resp.Warned != warned {
		t.Fatalf("Warned = %v, want %v; response: %#v", resp.Warned, warned, resp)
	}

	if resp.Failed != failed {
		t.Fatalf("Failed = %v, want %v; response: %#v", resp.Failed, failed, resp)
	}

	if resp.Errored != errored {
		t.Fatalf("Errored = %v, want %v; response: %#v", resp.Errored, errored, resp)
	}
}
