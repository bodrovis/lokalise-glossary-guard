package validate

import "testing"

func TestColorHelpers_WithColor(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})

	noColor = false

	tests := []struct {
		name string
		fn   func(string) string
		want string
	}{
		{"green", green, clrGreen + "ok" + clrReset},
		{"red", red, clrRed + "ok" + clrReset},
		{"cyan", cyan, clrCyan + "ok" + clrReset},
		{"yellow", yellow, clrYellow + "ok" + clrReset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn("ok")
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorHelpers_NoColor(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})

	noColor = true

	tests := []struct {
		name string
		fn   func(string) string
	}{
		{"green", green},
		{"red", red},
		{"cyan", cyan},
		{"yellow", yellow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn("ok")
			if got != "ok" {
				t.Fatalf("got %q, want %q", got, "ok")
			}
		})
	}
}

func TestColorStatus_WithColor(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})

	noColor = false

	tests := []struct {
		status string
		want   string
	}{
		{"PASS", clrGreen + "PASS" + clrReset},
		{"WARN", clrYellow + "WARN" + clrReset},
		{"FAIL", clrRed + "FAIL" + clrReset},
		{"ERROR", clrRed + "ERROR" + clrReset},
		{"UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := colorStatus(tt.status)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorStatus_NoColor(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})

	noColor = true

	for _, status := range []string{"PASS", "WARN", "FAIL", "ERROR", "UNKNOWN"} {
		t.Run(status, func(t *testing.T) {
			got := colorStatus(status)
			if got != status {
				t.Fatalf("got %q, want %q", got, status)
			}
		})
	}
}
