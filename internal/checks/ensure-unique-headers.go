package checks

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

type ensureUniqueHeaders struct{}

const ensUniqHdrsName = "ensure-unique-headers"

func (ensureUniqueHeaders) Name() string   { return ensUniqHdrsName }
func (ensureUniqueHeaders) FailFast() bool { return false }
func (ensureUniqueHeaders) Priority() int  { return 102 }

func (ensureUniqueHeaders) Run(data []byte, _ string, _ []string) Result {
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = ';'
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true

	header, err := r.Read()
	if errors.Is(err, io.EOF) || len(header) == 0 {
		return Result{Name: ensUniqHdrsName, Status: Error, Message: "cannot read header: empty file"}
	}
	if err != nil {
		return Result{Name: ensUniqHdrsName, Status: Error, Message: fmt.Sprintf("cannot read header: %v", err)}
	}

	seen := make(map[string]int, len(header))
	dups := make([]string, 0, 4)

	normalize := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "-", "_")
		return s
	}

	for i, h := range header {
		v := strings.TrimSpace(h)
		if i == 0 {
			v = strings.TrimPrefix(v, "\uFEFF") // strip BOM if present
		}
		key := normalize(v)

		if prev, ok := seen[key]; ok {
			dups = append(dups, fmt.Sprintf("%q (columns %d and %d)", v, prev+1, i+1))
		} else {
			seen[key] = i
		}
	}

	if len(dups) > 0 {
		sort.Strings(dups)

		const maxShow = 10
		show := dups
		more := 0

		if len(dups) > maxShow {
			show = dups[:maxShow]
			more = len(dups) - maxShow
		}

		msg := "duplicate header name(s) found:\n  - " + strings.Join(show, "\n  - ")
		if more > 0 {
			msg += fmt.Sprintf("\n  ...and %d more", more)
		}

		return Result{Name: ensUniqHdrsName, Status: Fail, Message: msg}
	}

	return Result{Name: ensUniqHdrsName, Status: Pass, Message: "All header names are unique"}
}

func init() { Register(ensureUniqueHeaders{}) }
