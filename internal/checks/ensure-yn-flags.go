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

type ensureYNFlags struct{}

const ensYesNoFlagsName = "ensure-yn-flags"

func (ensureYNFlags) Name() string   { return ensYesNoFlagsName }
func (ensureYNFlags) FailFast() bool { return false }
func (ensureYNFlags) Priority() int  { return 100 }

func (ensureYNFlags) Run(data []byte, _ string, _ []string) Result {
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = ';'
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	r.ReuseRecord = true

	header, err := r.Read()
	if errors.Is(err, io.EOF) || len(header) == 0 {
		return Result{Name: ensYesNoFlagsName, Status: Error, Message: "cannot read header: empty file"}
	}
	if err != nil {
		return Result{Name: ensYesNoFlagsName, Status: Error, Message: fmt.Sprintf("cannot read header: %v", err)}
	}

	// normalize header (lowercase, trim, strip BOM on first)
	norm := make([]string, len(header))
	for i, h := range header {
		v := strings.TrimSpace(h)
		if i == 0 {
			v = strings.TrimPrefix(v, "\uFEFF")
		}
		norm[i] = strings.ToLower(v)
	}

	idxs := make(map[string]int, 3)
	for i, col := range norm {
		switch col {
		case "casesensitive", "translatable", "forbidden":
			idxs[col] = i
		}
	}

	if len(idxs) == 0 {
		return Result{Name: ensYesNoFlagsName, Status: Pass, Message: "No Y/N flag columns present (nothing to validate)"}
	}

	isYN := func(s string) bool {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "yes", "no":
			return true
		default:
			return false
		}
	}

	type bad struct {
		line int
		col  string
		val  string
	}
	var bads []bad

	line := 1 // header line
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		line++
		if err != nil {
			return Result{Name: ensYesNoFlagsName, Status: Error, Message: fmt.Sprintf("CSV parse error at line %d: %v", line, err)}
		}

		for colName, idx := range idxs {
			val := ""
			if idx < len(rec) {
				val = rec[idx]
			}
			if strings.TrimSpace(val) == "" || !isYN(val) {
				bads = append(bads, bad{line: line, col: colName, val: strings.TrimSpace(val)})
			}
		}
	}

	if len(bads) > 0 {
		sort.Slice(bads, func(i, j int) bool {
			if bads[i].col != bads[j].col {
				return bads[i].col < bads[j].col
			}
			return bads[i].line < bads[j].line
		})

		const maxShow = 10
		show := bads
		more := 0

		if len(bads) > maxShow {
			show = bads[:maxShow]
			more = len(bads) - maxShow
		}

		var b strings.Builder
		b.WriteString("invalid values in Y/N flag columns (expected 'yes' or 'no'):")
		for _, it := range show {
			v := it.val
			if v == "" {
				v = "<empty>"
			}
			fmt.Fprintf(&b, "\n  - line %d, column %s: %q", it.line, it.col, v)
		}

		if more > 0 {
			fmt.Fprintf(&b, "\n  ...and %d more", more)
		}

		return Result{Name: ensYesNoFlagsName, Status: Fail, Message: b.String()}
	}

	return Result{Name: ensYesNoFlagsName, Status: Pass, Message: "Y/N flag columns valid (yes/no only)"}
}

func init() { Register(ensureYNFlags{}) }
