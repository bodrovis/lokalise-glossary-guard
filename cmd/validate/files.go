package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

func expandFiles(fs []string) ([]string, error) {
	seen := map[string]struct{}{}
	var out []string

	for _, f := range fs {
		for _, raw := range strings.Split(f, ",") {
			p := strings.TrimSpace(raw)
			if p == "" {
				continue
			}
			if hasGlob(p) {
				matches, err := filepath.Glob(p)
				if err != nil {
					return nil, err
				}
				for _, m := range matches {
					info, err := os.Stat(m)
					if err == nil && !info.IsDir() {
						if _, ok := seen[m]; !ok {
							seen[m] = struct{}{}
							out = append(out, m)
						}
					}
				}
				continue
			}
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no files matched the provided patterns")
	}
	return out, nil
}

func writeFixedFileIfNeeded(
	b *strings.Builder,
	opts checks.RunOptions,
	resp guard.ValidateResponse,
) error {
	if opts.FixMode == checks.FixNone || !resp.Fixed {
		return nil
	}

	outPath := withFixedPostfix(resp.Summary.FinalPath)

	if err := os.WriteFile(outPath, resp.FixedData, 0o644); err != nil {
		fmt.Fprintf(b, "%s writing fixed file: %v\n", red("ERROR"), err)
		return err
	}

	fmt.Fprintf(b, "%s wrote fixed file: %s (bytes=%d)\n", cyan("Info"), outPath, len(resp.FixedData))
	return nil
}

func withFixedPostfix(p string) string {
	ext := filepath.Ext(p)
	base := strings.TrimSuffix(p, ext)
	if strings.HasSuffix(base, "_fixed") {
		return base + ext
	}
	return base + "_fixed" + ext
}

func hasGlob(s string) bool { return strings.ContainsAny(s, "*?[]") }
