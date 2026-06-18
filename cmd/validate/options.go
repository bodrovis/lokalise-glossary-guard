package validate

import (
	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

func buildRunOptions() checks.RunOptions {
	fm := checks.FixNone
	if doFix {
		fm = checks.FixIfNotPass
	}
	return checks.RunOptions{
		FixMode:       fm,
		RerunAfterFix: rerunAfterFix,
		HardFailOnErr: hardFailOnErr,
	}
}
