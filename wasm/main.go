//go:build js && wasm

package main

import (
	"syscall/js"
)

func main() {
	js.Global().Set("validateGlossaryGuard", js.FuncOf(validateGlossaryGuard))

	// Keep WASM alive.
	select {}
}
