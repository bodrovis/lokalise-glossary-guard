package validate

const (
	clrReset  = "\x1b[0m"
	clrRed    = "\x1b[31m"
	clrGreen  = "\x1b[32m"
	clrYellow = "\x1b[33m"
	clrCyan   = "\x1b[36m"
)

type colorizer struct {
	disabled bool
}

func newColorizer(noColor bool) colorizer {
	return colorizer{
		disabled: noColor,
	}
}

func (c colorizer) colorize(s, color string) string {
	if c.disabled {
		return s
	}

	return color + s + clrReset
}

func (c colorizer) green(s string) string {
	return c.colorize(s, clrGreen)
}

func (c colorizer) red(s string) string {
	return c.colorize(s, clrRed)
}

func (c colorizer) cyan(s string) string {
	return c.colorize(s, clrCyan)
}

func (c colorizer) yellow(s string) string {
	return c.colorize(s, clrYellow)
}

func (c colorizer) status(s string) string {
	switch s {
	case "PASS":
		return c.green(s)
	case "WARN":
		return c.yellow(s)
	case "FAIL", "ERROR":
		return c.red(s)
	default:
		return s
	}
}
