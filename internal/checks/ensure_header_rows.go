package checks

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

type ensureHeader struct{}

func (ensureHeader) Name() string   { return "ensure-header-and-rows" }
func (ensureHeader) FailFast() bool { return true }
func (ensureHeader) Priority() int  { return 3 }

func (ensureHeader) Run(filePath string) Result {
	const checkName = "ensure-header-and-rows"

	f, err := os.Open(filePath)
	if err != nil {
		return Result{Name: checkName, Status: Error, Message: fmt.Sprintf("cannot open file: %v", err)}
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close %s: %v\n", filePath, cerr)
		}
	}()

	br := bufio.NewReader(f)

	rawHeader, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		return Result{Name: checkName, Status: Error, Message: fmt.Sprintf("read error: %v", err)}
	}
	headerLine := strings.TrimRight(rawHeader, "\r\n")
	if len(headerLine) == 0 {
		return Result{Name: checkName, Status: Fail, Message: "Empty file: header row is required"}
	}

	hasSemicolon := strings.Contains(headerLine, ";")
	hasComma := strings.Contains(headerLine, ",")
	hasTab := strings.Contains(headerLine, "\t")

	if !hasSemicolon {
		switch {
		case hasComma:
			return Result{Name: checkName, Status: Fail, Message: "Header appears to use ',' as delimiter. Expected ';'."}
		case hasTab:
			return Result{Name: checkName, Status: Fail, Message: "Header appears to use TAB as delimiter. Expected ';'."}
		default:
			return Result{Name: checkName, Status: Fail, Message: "Header missing semicolons — expected ';' as delimiter"}
		}
	}
	if hasComma || hasTab {
		return Result{Name: checkName, Status: Fail, Message: "Header uses mixed delimiters. Expected semicolons (';') only"}
	}

	hdrCSV := csv.NewReader(strings.NewReader(headerLine))
	hdrCSV.Comma = ';'
	hdrCSV.FieldsPerRecord = -1
	hdrCSV.LazyQuotes = false
	hdrCSV.TrimLeadingSpace = true

	header, err := hdrCSV.Read()
	if err != nil {
		return Result{Name: checkName, Status: Fail, Message: fmt.Sprintf("cannot parse header: %v", err)}
	}
	if len(header) < 2 {
		return Result{Name: checkName, Status: Fail, Message: "Malformed header: expected at least 2 semicolon-separated columns"}
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
		return Result{Name: checkName, Status: Fail, Message: fmt.Sprintf("Header missing required columns: %s", strings.Join(missing, ", "))}
	}

	if len(norm) < 2 || norm[0] != "term" || norm[1] != "description" {
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
			Name:   checkName,
			Status: Fail,
			Message: fmt.Sprintf(
				"Invalid header order: expected first two columns to be 'term;description', got '%s;%s' (found term at #%d, description at #%d)",
				got0, got1, termPos+1, descPos+1,
			),
		}
	}

	{
		f2, err := os.Open(filePath)
		if err == nil {
			defer func() {
				if cerr := f2.Close(); cerr != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to close %s: %v\n", filePath, cerr)
				}
			}()

			sc := bufio.NewScanner(f2)
			buf := make([]byte, 64*1024)
			sc.Buffer(buf, 10*1024*1024)

			lineNo := 0
			blankLines := make([]int, 0, 8)

			if sc.Scan() {
				lineNo++
			}
			for sc.Scan() {
				lineNo++
				if strings.TrimSpace(sc.Text()) == "" {
					blankLines = append(blankLines, lineNo)
				}
			}
			if err := sc.Err(); err != nil {
				return Result{Name: checkName, Status: Error, Message: fmt.Sprintf("scan error: %v", err)}
			}
			if len(blankLines) > 0 {
				const maxShow = 10
				list := blankLines
				more := 0
				if len(blankLines) > maxShow {
					list = blankLines[:maxShow]
					more = len(blankLines) - maxShow
				}
				msg := fmt.Sprintf("Blank lines are not allowed after header. Found at row(s): %s", intsToCSV(list))
				if more > 0 {
					msg += fmt.Sprintf(", …and %d more", more)
				}
				return Result{Name: checkName, Status: Fail, Message: msg}
			}
		}
	}

	r := csv.NewReader(br)
	r.Comma = ';'
	r.FieldsPerRecord = len(header)
	r.LazyQuotes = false
	r.TrimLeadingSpace = true

	seenValid := 0
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Result{Name: checkName, Status: Fail, Message: fmt.Sprintf("CSV parse error on data (check delimiter/quoting): %v", err)}
		}

		allEmpty := true
		for _, v := range rec {
			if strings.TrimSpace(v) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			return Result{Name: checkName, Status: Fail, Message: "Blank data row is not allowed"}
		}

		seenValid++
	}

	if seenValid == 0 {
		return Result{Name: checkName, Status: Fail, Message: "No data rows found after header"}
	}

	return Result{Name: checkName, Status: Pass, Message: "Header valid; required columns present; ';' delimiter confirmed; no blank lines; data parsed successfully"}
}

func init() { Register(ensureHeader{}) }

func intsToCSV(xs []int) string {
	s := make([]string, len(xs))
	for i, v := range xs {
		s[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(s, ", ")
}
