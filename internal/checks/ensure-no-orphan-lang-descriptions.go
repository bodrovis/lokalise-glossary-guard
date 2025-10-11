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

type ensureNoOrphanLangDescriptions struct{}

const ensNoOrphanDescName = "ensure-no-orphan-lang-descriptions"

func (ensureNoOrphanLangDescriptions) Name() string   { return ensNoOrphanDescName }
func (ensureNoOrphanLangDescriptions) FailFast() bool { return false }
func (ensureNoOrphanLangDescriptions) Priority() int  { return 100 }

func (ensureNoOrphanLangDescriptions) Run(data []byte, _ string, _ []string) Result {
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = ';'
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true

	header, err := r.Read()
	if errors.Is(err, io.EOF) || len(header) == 0 {
		return Result{Name: ensNoOrphanDescName, Status: Error, Message: "cannot read header: empty file"}
	}
	if err != nil {
		return Result{Name: ensNoOrphanDescName, Status: Error, Message: fmt.Sprintf("cannot read header: %v", err)}
	}
	if len(header) < 2 {
		return Result{Name: ensNoOrphanDescName, Status: Error, Message: "malformed header: expected at least 2 columns"}
	}

	baseLangs := make(map[string]struct{}, len(header))

	var orphans []string

	// skip term;description
	for _, raw := range header[2:] {
		raw = strings.TrimSpace(raw)
		lc := strings.ToLower(raw)
		if lc == "" {
			continue
		}

		if strings.HasSuffix(lc, "_description") {
			base := strings.TrimSuffix(lc, "_description")

			normBase := normalizeLang(base)

			if _, ok := baseLangs[normBase]; !ok {
				orphans = append(orphans, raw)
			}
			continue
		}

		baseLangs[normalizeLang(lc)] = struct{}{}
	}

	if len(orphans) > 0 {
		sort.Strings(orphans)

		return Result{
			Name:   ensNoOrphanDescName,
			Status: Fail,
			Message: fmt.Sprintf(
				"found orphan description column(s) without a matching language column: %s",
				strings.Join(orphans, ", "),
			),
		}
	}

	return Result{Name: ensNoOrphanDescName, Status: Pass, Message: "No orphan *_description columns found"}
}

func init() { Register(ensureNoOrphanLangDescriptions{}) }
