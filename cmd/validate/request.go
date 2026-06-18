package validate

import (
	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

func buildValidateRequest(path string, data []byte, langs []string, opts checks.RunOptions) guard.ValidateRequest {
	rerunAfterFix := opts.RerunAfterFix

	return guard.ValidateRequest{
		Path:          path,
		Data:          data,
		Langs:         langs,
		Fix:           opts.FixMode != checks.FixNone,
		RerunAfterFix: &rerunAfterFix,
		HardFailOnErr: opts.HardFailOnErr,
	}
}
