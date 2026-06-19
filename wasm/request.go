//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

func validateRequestFromArgs(args []js.Value) (guard.ValidateRequest, error) {
	var req guard.ValidateRequest

	if len(args) == 0 {
		return req, fmt.Errorf("missing input")
	}

	input, err := inputToJSON(args[0])
	if err != nil {
		return req, err
	}

	if strings.TrimSpace(input) == "null" {
		return req, fmt.Errorf("input must be a JSON string or object")
	}

	if err := json.Unmarshal([]byte(input), &req); err != nil {
		return req, fmt.Errorf(
			"invalid input json: %v; pass an object like {data: csvText} or a JSON string",
			err,
		)
	}

	return req, nil
}

func inputToJSON(v js.Value) (string, error) {
	if v.IsUndefined() || v.IsNull() {
		return "", fmt.Errorf("input must be a JSON string or object")
	}

	if v.Type() == js.TypeString {
		return v.String(), nil
	}

	jsonValue := js.Global().Get("JSON").Call("stringify", v)
	if jsonValue.IsUndefined() || jsonValue.IsNull() {
		return "", fmt.Errorf("input must be a JSON string or object")
	}

	return jsonValue.String(), nil
}
