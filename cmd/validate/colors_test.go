package validate

import "testing"

func TestColorizer_WithColor(t *testing.T) {
	t.Parallel()

	c := newColorizer(false)

	tests := []struct {
		name string
		fn   func(string) string
		want string
	}{
		{"green", c.green, clrGreen + "ok" + clrReset},
		{"red", c.red, clrRed + "ok" + clrReset},
		{"cyan", c.cyan, clrCyan + "ok" + clrReset},
		{"yellow", c.yellow, clrYellow + "ok" + clrReset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.fn("ok")

			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorizer_Disabled(t *testing.T) {
	t.Parallel()

	c := newColorizer(true)

	tests := []struct {
		name string
		fn   func(string) string
	}{
		{"green", c.green},
		{"red", c.red},
		{"cyan", c.cyan},
		{"yellow", c.yellow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.fn("ok")

			if got != "ok" {
				t.Fatalf("got %q, want %q", got, "ok")
			}
		})
	}
}

func TestColorizer_Status(t *testing.T) {
	t.Parallel()

	t.Run("with color", func(t *testing.T) {
		t.Parallel()

		c := newColorizer(false)

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
				t.Parallel()

				got := c.status(tt.status)

				if got != tt.want {
					t.Fatalf("got %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		c := newColorizer(true)

		for _, status := range []string{
			"PASS",
			"WARN",
			"FAIL",
			"ERROR",
			"UNKNOWN",
		} {
			t.Run(status, func(t *testing.T) {
				t.Parallel()

				got := c.status(status)

				if got != status {
					t.Fatalf("got %q, want %q", got, status)
				}
			})
		}
	})
}
