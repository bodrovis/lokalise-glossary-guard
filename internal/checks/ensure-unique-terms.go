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

type ensureUniqueTermsCS struct{}

const ensUniqTermsName = "ensure-unique-terms"

func (ensureUniqueTermsCS) Name() string   { return ensUniqTermsName }
func (ensureUniqueTermsCS) FailFast() bool { return false }
func (ensureUniqueTermsCS) Priority() int  { return 100 }

func (ensureUniqueTermsCS) Run(data []byte, _ string, _ []string) Result {
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = ';'
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	r.ReuseRecord = true

	header, err := r.Read()
	if errors.Is(err, io.EOF) || len(header) == 0 {
		return Result{Name: ensUniqTermsName, Status: Error, Message: "cannot read header: empty file"}
	}
	if err != nil {
		return Result{Name: ensUniqTermsName, Status: Error, Message: fmt.Sprintf("cannot read header: %v", err)}
	}

	termIdx := -1
	for i, h := range header {
		v := strings.TrimSpace(h)
		if i == 0 {
			v = strings.TrimPrefix(v, "\uFEFF")
		}
		if strings.EqualFold(v, "term") {
			termIdx = i
			break
		}
	}

	if termIdx == -1 {
		return Result{Name: ensUniqTermsName, Status: Error, Message: "header does not contain 'term' column"}
	}

	type hit struct {
		firstLine int
		line      int
		value     string
	}

	seen := make(map[string]int, 1024) // term -> first line
	var dups []hit

	line := 1 // header = line 1
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		line++
		if err != nil {
			return Result{Name: ensUniqTermsName, Status: Error, Message: fmt.Sprintf("csv parse error at line %d: %v", line, err)}
		}
		if termIdx >= len(rec) {
			continue
		}
		val := strings.TrimSpace(rec[termIdx])
		if val == "" {
			continue
		}

		if line == 2 {
			val = strings.TrimPrefix(val, "\uFEFF")
		}

		if first, ok := seen[val]; ok {
			dups = append(dups, hit{firstLine: first, line: line, value: val})
		} else {
			seen[val] = line
		}
	}

	if len(dups) > 0 {
		sort.Slice(dups, func(i, j int) bool {
			if dups[i].value != dups[j].value {
				return dups[i].value < dups[j].value
			}
			if dups[i].firstLine != dups[j].firstLine {
				return dups[i].firstLine < dups[j].firstLine
			}
			return dups[i].line < dups[j].line
		})

		const maxShow = 10
		show := dups
		more := 0

		if len(dups) > maxShow {
			show = dups[:maxShow]
			more = len(dups) - maxShow
		}

		var b strings.Builder
		b.WriteString("duplicate terms (case-sensitive, trimmed) detected:\n")

		for _, h := range show {
			fmt.Fprintf(&b, "  - %q at lines %d and %d\n", h.value, h.firstLine, h.line)
		}

		if more > 0 {
			fmt.Fprintf(&b, "  …and %d more", more)
		}

		return Result{Name: ensUniqTermsName, Status: Fail, Message: b.String()}
	}

	return Result{Name: ensUniqTermsName, Status: Pass, Message: "All terms are unique (case-sensitive)"}
}

func init() { Register(ensureUniqueTermsCS{}) }
