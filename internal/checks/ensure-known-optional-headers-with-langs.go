package checks

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

type ensureKnownHeadersWithLangs struct{}

const ensKnownHdrsName = "ensure-known-optional-headers-with-langs"

func (ensureKnownHeadersWithLangs) Name() string   { return ensKnownHdrsName }
func (ensureKnownHeadersWithLangs) FailFast() bool { return false }
func (ensureKnownHeadersWithLangs) Priority() int  { return 100 }

var (
	reLang       = regexp.MustCompile(`^[A-Za-z]{2,3}(?:[_-][A-Za-z0-9]{2,8})*$`)
	reLangDesc   = regexp.MustCompile(`^([A-Za-z]{2,3}(?:[_-][A-Za-z0-9]{2,8})*)_description$`)
	allowedFixed = map[string]struct{}{
		"casesensitive": {},
		"translatable":  {},
		"forbidden":     {},
		"tags":          {},
	}
)

func (ensureKnownHeadersWithLangs) Run(data []byte, _ string, langs []string) Result {
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = ';'
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	r.ReuseRecord = true

	header, err := r.Read()

	if errors.Is(err, io.EOF) || len(header) == 0 {
		return Result{Name: ensKnownHdrsName, Status: Error, Message: "cannot read header: empty file"}
	}

	if err != nil {
		return Result{Name: ensKnownHdrsName, Status: Error, Message: fmt.Sprintf("cannot read header: %v", err)}
	}

	norm := make([]string, len(header))
	for i, h := range header {
		v := strings.TrimSpace(h)
		if i == 0 {
			v = strings.TrimPrefix(v, "\uFEFF") // drop BOM
		}
		norm[i] = v
	}
	if len(norm) < 2 {
		return Result{Name: ensKnownHdrsName, Status: Error, Message: "malformed header: expected at least 2 columns"}
	}

	declared := makeLangSet(langs...)
	haveLangs := len(declared) > 0

	var unknown []string
	var looksLikeLang []string

	// validate columns after the first two
	for _, col := range norm[2:] {
		raw := col
		lc := strings.ToLower(raw)

		if _, ok := allowedFixed[lc]; ok {
			continue
		}

		if haveLangs {
			base := lc
			base = strings.TrimSuffix(base, "_description")

			if _, ok := declared[normalizeLang(base)]; ok {
				continue
			}

			unknown = append(unknown, raw)

			continue
		}

		if reLang.MatchString(raw) || reLangDesc.MatchString(lc) {
			looksLikeLang = append(looksLikeLang, raw)
			continue
		}

		unknown = append(unknown, raw)
	}

	if haveLangs {
		var missing []string

		presentLangs := map[string]struct{}{}

		for _, col := range norm[2:] {
			lc := strings.ToLower(col)
			if strings.HasSuffix(lc, "_description") {
				continue
			}
			presentLangs[normalizeLang(lc)] = struct{}{}
		}

		for lang := range declared {
			if _, ok := presentLangs[lang]; !ok {
				missing = append(missing, lang)
			}
		}

		sort.Strings(missing)
		sort.Strings(unknown)

		if len(missing) > 0 {
			return Result{
				Name:   ensKnownHdrsName,
				Status: Fail,
				Message: fmt.Sprintf(
					"missing required language column(s): %s",
					strings.Join(missing, ", "),
				),
			}
		}

		if len(unknown) > 0 {
			return Result{
				Name:   ensKnownHdrsName,
				Status: Fail,
				Message: fmt.Sprintf(
					"unsupported header column(s): %s. Allowed after 'term;description' are: casesensitive, translatable, forbidden, tags, declared language codes (%s) and their *_description.",
					strings.Join(unknown, ", "),
					strings.Join(sortedKeys(declared), ", "),
				),
			}
		}

		return Result{Name: ensKnownHdrsName, Status: Pass, Message: "Optional header columns are valid for declared languages"}
	}

	sort.Strings(looksLikeLang)
	sort.Strings(unknown)

	if len(looksLikeLang) > 0 {
		return Result{
			Name:   ensKnownHdrsName,
			Status: Warn,
			Message: fmt.Sprintf(
				"header contains language-like column(s) but no languages were provided: %s. Pass languages (e.g. --langs en,fr,de) or remove unsupported columns.",
				strings.Join(looksLikeLang, ", "),
			),
		}
	}

	if len(unknown) > 0 {
		return Result{
			Name:   ensKnownHdrsName,
			Status: Fail,
			Message: fmt.Sprintf(
				"unsupported header column(s): %s. Allowed after 'term;description' are: casesensitive, translatable, forbidden, tags, language ISO codes and their *_description.",
				strings.Join(unknown, ", "),
			),
		}
	}

	return Result{Name: ensKnownHdrsName, Status: Pass, Message: "Optional header columns are valid"}
}

func init() { Register(ensureKnownHeadersWithLangs{}) }

// helpers

func normalizeLang(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

func makeLangSet(langs ...string) map[string]struct{} {
	if len(langs) == 0 {
		return nil
	}

	m := make(map[string]struct{}, len(langs))
	for _, l := range langs {
		n := normalizeLang(l)
		if n != "" {
			m[n] = struct{}{}
		}
	}

	return m
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	sort.Strings(out)

	return out
}
