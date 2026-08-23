package guard

import (
	"context"
	"encoding/json/v2"
	"errors"

	_ "github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks/all"
	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/validator"
)

func ValidateBytesJSON(ctx context.Context, req ValidateRequest) ([]byte, error) {
	resp, err := ValidateBytes(ctx, req)

	if err != nil && errors.Is(err, context.Canceled) {
		return nil, err
	}

	return json.Marshal(resp)
}

func ValidateBytes(ctx context.Context, req ValidateRequest) (ValidateResponse, error) {
	opts := runOptions(req)
	data := requestData(req)
	langs := PreprocessLangs(req.Langs)

	coreSummary, err := validator.Validate(ctx, req.Path, data, langs, opts)
	resp := responseFromSummary(req.Path, coreSummary, err)

	return resp, err
}
