package checks

import "unicode/utf8"

type ensureUTF8 struct{}

const ensUTF8Name = "ensure-utf8-encoding"

func (ensureUTF8) Name() string   { return ensUTF8Name }
func (ensureUTF8) FailFast() bool { return true }
func (ensureUTF8) Priority() int  { return 2 }

func (ensureUTF8) Run(data []byte, _filePath string, _langs []string) Result {
	if len(data) == 0 {
		return Result{Name: ensUTF8Name, Status: Fail, Message: "Empty file: cannot determine encoding"}
	}

	if utf8.Valid(data) {
		return Result{Name: ensUTF8Name, Status: Pass, Message: "File encoding is valid UTF-8"}
	}

	i := 0
	for i < len(data) {
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size == 1 {
			break
		}
		i += size
	}

	return Result{
		Name:    ensUTF8Name,
		Status:  Fail,
		Message: formatUTF8Error(i, len(data)),
	}
}

func formatUTF8Error(pos, total int) string {
	if pos >= total {
		return "File encoding is not valid UTF-8 (corruption near end of file)"
	}
	return "File encoding is not valid UTF-8 (invalid byte sequence near offset " + itoa(pos) + ")"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func init() { Register(ensureUTF8{}) }
