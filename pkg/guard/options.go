package guard

import "github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"

func runOptions(req ValidateRequest) checks.RunOptions {
	fixMode := checks.FixNone
	if req.Fix {
		fixMode = checks.FixIfNotPass
	}

	return checks.RunOptions{
		FixMode:       fixMode,
		RerunAfterFix: rerunAfterFix(req),
		HardFailOnErr: req.HardFailOnErr,
	}
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
