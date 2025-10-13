package checks

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
)

type ensureHeader struct{}

var ensHdrRowsName = "ensure-header-and-rows"

func (ensureHeader) Name() string   { return ensHdrRowsName }
func (ensureHeader) FailFast() bool { return true }
func (ensureHeader) Priority() int  { return 3 }

func (ensureHeader) Run(data []byte, _filePath string, _langs []string) Result {
	if len(data) == 0 {
		return Result{Name: ensHdrRowsName, Status: Fail, Message: "Empty file: header row is required"}
	}

	br := bufio.NewReader(bytes.NewReader(data))

	rawHeader, err := br.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return Result{Name: ensHdrRowsName, Status: Error, Message: fmt.Sprintf("read error: %v", err)}
	}

	headerLine := strings.TrimRight(rawHeader, "\r\n")
	if headerLine == "" {
		return Result{Name: ensHdrRowsName, Status: Fail, Message: "Empty file: header row is required"}
	}

	hasSemicolon := strings.Contains(headerLine, ";")
	hasComma := strings.Contains(headerLine, ",")
	hasTab := strings.Contains(headerLine, "\t")

	if !hasSemicolon {
		switch {
		case hasComma:
			return Result{Name: ensHdrRowsName, Status: Fail, Message: "Header appears to use ',' as delimiter. Expected ';'"}
		case hasTab:
			return Result{Name: ensHdrRowsName, Status: Fail, Message: "Header appears to use TAB as delimiter. Expected ';'"}
		default:
			return Result{Name: ensHdrRowsName, Status: Fail, Message: "Header missing semicolons. Expected ';' as delimiter"}
		}
	}

	if hasComma || hasTab {
		return Result{Name: ensHdrRowsName, Status: Fail, Message: "Header uses mixed delimiters. Expected semicolons (';') only"}
	}

	hdrCSV := csv.NewReader(strings.NewReader(headerLine))
	hdrCSV.Comma = ';'
	hdrCSV.FieldsPerRecord = -1
	hdrCSV.LazyQuotes = false
	hdrCSV.TrimLeadingSpace = true

	header, err := hdrCSV.Read()
	if err != nil {
		return Result{Name: ensHdrRowsName, Status: Fail, Message: fmt.Sprintf("cannot parse header: %v", err)}
	}
	if len(header) < 2 {
		return Result{Name: ensHdrRowsName, Status: Fail, Message: "Malformed header: expected at least 2 semicolon-separated columns"}
	}

	norm := make([]string, len(header))
	for i, h := range header {
		n := strings.ToLower(strings.TrimSpace(h))
		if i == 0 {
			n = strings.TrimPrefix(n, "\uFEFF")
		}
		norm[i] = n
	}

	req := map[string]bool{"term": false, "description": false}
	for _, col := range norm {
		if _, ok := req[col]; ok {
			req[col] = true
		}
	}
	var missing []string
	for k, ok := range req {
		if !ok {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return Result{Name: ensHdrRowsName, Status: Fail, Message: fmt.Sprintf("Header missing required columns: %s", strings.Join(missing, ", "))}
	}

	if norm[0] != "term" || norm[1] != "description" {
		termPos, descPos := -1, -1

		for i, c := range norm {
			if c == "term" && termPos == -1 {
				termPos = i
			}
			if c == "description" && descPos == -1 {
				descPos = i
			}
		}

		got0, got1 := "", ""
		if len(norm) > 0 {
			got0 = norm[0]
		}
		if len(norm) > 1 {
			got1 = norm[1]
		}

		return Result{
			Name:   ensHdrRowsName,
			Status: Fail,
			Message: fmt.Sprintf(
				"Invalid header order: expected first two columns to be 'term;description', got '%s;%s' (found term at #%d, description at #%d)",
				got0, got1, termPos+1, descPos+1,
			),
		}
	}

	{
		sc := bufio.NewScanner(bytes.NewReader(data))
		buf := make([]byte, 64*1024)
		sc.Buffer(buf, 10*1024*1024)

		lineNo := 0
		if sc.Scan() {
			lineNo++
		}
		for sc.Scan() {
			lineNo++
			if strings.TrimSpace(sc.Text()) == "" {
				return Result{
					Name:    ensHdrRowsName,
					Status:  Warn,
					Message: fmt.Sprintf("Blank data row might cause issues, better remove (line %d)", lineNo),
				}
			}
		}
		if err := sc.Err(); err != nil {
			return Result{Name: ensHdrRowsName, Status: Error, Message: fmt.Sprintf("scan error: %v", err)}
		}
	}

	r := csv.NewReader(br)
	r.Comma = ';'
	r.FieldsPerRecord = len(header)
	r.LazyQuotes = false
	r.TrimLeadingSpace = true
	r.ReuseRecord = true

	seenValid := 0

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			var pe *csv.ParseError
			if errors.As(err, &pe) {
				fileLine := 1 + pe.Line
				switch pe.Err {
				case csv.ErrFieldCount:
					return Result{
						Name:   ensHdrRowsName,
						Status: Fail,
						Message: fmt.Sprintf(
							"CSV parse error at line %d: wrong number of fields (expected %d)",
							fileLine, len(header),
						),
					}
				default:
					return Result{
						Name:   ensHdrRowsName,
						Status: Fail,
						Message: fmt.Sprintf(
							"CSV parse error at line %d: %v",
							fileLine, pe.Err,
						),
					}
				}
			}
			return Result{
				Name:    ensHdrRowsName,
				Status:  Fail,
				Message: fmt.Sprintf("CSV parse error on data (check delimiter/quoting): %v", err),
			}
		}

		allEmpty := true
		for _, v := range rec {
			if strings.TrimSpace(v) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			line := 1 + seenValid + 1 // header(1) + already valid rows + current
			return Result{Name: ensHdrRowsName, Status: Warn, Message: fmt.Sprintf("Blank data row might cause issues, better remove (line %d)", line)}
		}

		seenValid++
	}

	if seenValid == 0 {
		return Result{Name: ensHdrRowsName, Status: Fail, Message: "No data rows found after header"}
	}

	return Result{
		Name:    ensHdrRowsName,
		Status:  Pass,
		Message: "Header valid; required columns present; ';' delimiter confirmed; data parsed successfully",
	}
}

func init() { Register(ensureHeader{}) }
