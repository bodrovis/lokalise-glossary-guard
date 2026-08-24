package guard

import (
	"slices"
	"strings"
)

func PreprocessLangs(ls []string) []string {
	if len(ls) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(ls))
	out := make([]string, 0, len(ls))

	for _, v := range ls {
		for part := range strings.SplitSeq(v, ",") {
			s := strings.TrimSpace(part)
			if s == "" {
				continue
			}

			if _, ok := seen[s]; ok {
				continue
			}

			seen[s] = struct{}{}
			out = append(out, s)
		}
	}

	slices.Sort(out)

	return out
}
