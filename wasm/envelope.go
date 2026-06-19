//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

type wasmValidateEnvelope struct {
	OK     bool                    `json:"ok"`
	Result *guard.ValidateResponse `json:"result,omitempty"`
	Error  string                  `json:"error,omitempty"`
}

func encodeEnvelope(resp wasmValidateEnvelope) string {
	data, err := json.Marshal(resp)
	if err != nil {
		fallback, _ := json.Marshal(errorEnvelope(fmt.Sprintf("failed to encode response: %v", err)))
		return string(fallback)
	}

	return string(data)
}

func errorEnvelope(msg string) wasmValidateEnvelope {
	return wasmValidateEnvelope{
		OK:    false,
		Error: msg,
	}
}
