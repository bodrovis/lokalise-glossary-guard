package guard

import (
	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

func runOptions(req ValidateRequest) checks.RunOptions {
	opts := checks.RunOptions{
		FixMode:       checks.FixNone,
		RerunAfterFix: rerunAfterFix(req),
		HardFailOnErr: req.HardFailOnErr,
	}

	if req.Fix {
		opts.FixMode = checks.FixIfNotPass
	}

	return opts
}

func requestData(req ValidateRequest) []byte {
	if req.Data != nil {
		return req.Data
	}

	return []byte(req.Text)
}

func rerunAfterFix(req ValidateRequest) bool {
	if req.RerunAfterFix == nil {
		return true
	}

	return *req.RerunAfterFix
}
