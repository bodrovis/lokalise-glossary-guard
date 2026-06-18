package validate

var (
	clrReset  = "\x1b[0m"
	clrRed    = "\x1b[31m"
	clrGreen  = "\x1b[32m"
	clrYellow = "\x1b[33m"
	clrCyan   = "\x1b[36m"
)

func green(s string) string {
	if noColor {
		return s
	}
	return clrGreen + s + clrReset
}

func red(s string) string {
	if noColor {
		return s
	}
	return clrRed + s + clrReset
}

func cyan(s string) string {
	if noColor {
		return s
	}
	return clrCyan + s + clrReset
}

func yellow(s string) string {
	if noColor {
		return s
	}
	return clrYellow + s + clrReset
}

func colorStatus(s string) string {
	switch s {
	case "PASS":
		return green(s)
	case "WARN":
		return yellow(s)
	default:
		return red(s) // FAIL/ERROR
	}
}
