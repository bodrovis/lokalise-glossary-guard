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
	seen := make(map[string]struct{})
	var out []string

	for _, f := range fs {
		for raw := range strings.SplitSeq(f, ",") {
			p := strings.TrimSpace(raw)
			if p == "" {
				continue
			}

			paths, err := expandFilePattern(p)
			if err != nil {
				return nil, err
			}

			for _, path := range paths {
				if _, ok := seen[path]; ok {
					continue
				}

				seen[path] = struct{}{}
				out = append(out, path)
			}
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no files matched the provided patterns")
	}

	return out, nil
}

func expandFilePattern(pattern string) ([]string, error) {
	if !hasGlob(pattern) {
		return []string{pattern}, nil
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(matches))

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			return nil, fmt.Errorf("stat matched file %q: %w", match, err)
		}

		if info.IsDir() {
			continue
		}

		out = append(out, match)
	}

	return out, nil
}

func writeFixedFileIfNeeded(
	b *strings.Builder,
	opts checks.RunOptions,
	resp guard.ValidateResponse,
	colors colorizer,
) error {
	if opts.FixMode == checks.FixNone || !resp.Fixed {
		return nil
	}

	outPath, err := fixedOutputPath(resp)
	if err != nil {
		fmt.Fprintf(
			b,
			"%s writing fixed file: %v\n",
			colors.red("ERROR"),
			err,
		)
		return err
	}

	if err := os.WriteFile(outPath, resp.FixedData, 0o644); err != nil {
		fmt.Fprintf(
			b,
			"%s writing fixed file: %v\n",
			colors.red("ERROR"),
			err,
		)
		return err
	}

	fmt.Fprintf(
		b,
		"%s wrote fixed file: %s (bytes=%d)\n",
		colors.cyan("Info"),
		outPath,
		len(resp.FixedData),
	)
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

func fixedOutputPath(resp guard.ValidateResponse) (string, error) {
	p := resp.Summary.FinalPath
	if p == "" {
		p = resp.Path
	}
	if p == "" {
		return "", fmt.Errorf("fixed output path is empty")
	}
	return withFixedPostfix(p), nil
}
