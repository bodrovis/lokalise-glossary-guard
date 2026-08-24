package validate

import (
	"context"
	"errors"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

type job struct {
	index int
	path  string
}

type fileRunConfig struct {
	Langs       []string
	Separator   string
	MaxParallel uint
	Options     checks.RunOptions
	Colors      colorizer
}

type validateRunConfig struct {
	Files       []string
	Langs       []string
	MaxParallel uint
	Options     checks.RunOptions
	JSONOut     bool
	NoColor     bool
	Out         io.Writer
	ErrOut      io.Writer
}

func runValidate(ctx context.Context, cfg validateRunConfig) error {
	start := time.Now()
	colors := newColorizer(cfg.NoColor)

	fileCfg := fileRunConfig{
		Langs:       cfg.Langs,
		Separator:   strings.Repeat("─", 72),
		MaxParallel: cfg.MaxParallel,
		Options:     cfg.Options,
		Colors:      colors,
	}

	outcomes := fileCfg.runFiles(ctx, cfg.Files)

	return newFinalizer(
		cfg.Out,
		cfg.ErrOut,
		cfg.JSONOut,
		colors,
	).finalize(
		outcomes,
		len(cfg.Files),
		start,
	)
}

func (cfg fileRunConfig) runFiles(
	ctx context.Context,
	files []string,
) []fileOutcome {
	outcomes := make([]fileOutcome, len(files))
	jobs := make(chan job)

	var wg sync.WaitGroup

	for range workerCount(cfg.MaxParallel, len(files)) {
		wg.Go(func() {
			for j := range jobs {
				outcomes[j.index] = cfg.runOneFile(
					ctx,
					j.index,
					j.path,
				)
			}
		})
	}

	for i, path := range files {
		select {
		case <-ctx.Done():
			cfg.markCanceledFiles(
				outcomes,
				files[i:],
				i,
				ctx.Err(),
			)

			close(jobs)
			wg.Wait()

			return outcomes

		case jobs <- job{
			index: i,
			path:  path,
		}:
		}
	}

	close(jobs)
	wg.Wait()

	return outcomes
}

func workerCount(maxParallel uint, filesCount int) int {
	if maxParallel < 1 {
		maxParallel = uint(runtime.GOMAXPROCS(0))
	}

	if filesCount < 1 {
		return 1
	}

	return int(min(maxParallel, uint(filesCount)))
}

func (cfg fileRunConfig) runOneFile(
	ctx context.Context,
	index int,
	path string,
) fileOutcome {
	var b strings.Builder
	renderer := newFileRenderer(&b, cfg)

	renderer.fileHeader(index, path)

	resp, err := cfg.validateFile(ctx, path)
	if err != nil {
		return cfg.opErrorOutcome(
			index,
			path,
			&b,
			err,
		)
	}

	oc := fileOutcome{
		Idx:  index,
		Path: path,
	}

	applyGuardResponse(&oc, resp)

	renderer.validationReport(
		path,
		resp.Summary,
	)

	if err := writeFixedFileIfNeeded(
		&b,
		cfg.Options,
		resp,
		cfg.Colors,
	); err != nil {
		oc.HadOpErr = true
		oc.Errored++
	}

	renderer.result(resp)

	b.WriteString(cfg.Separator)
	b.WriteByte('\n')

	oc.Output = b.String()

	return oc
}

func (cfg fileRunConfig) validateFile(
	ctx context.Context,
	path string,
) (guard.ValidateResponse, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return guard.ValidateResponse{}, err
	}

	resp, err := guard.ValidateBytes(
		ctx,
		buildValidateRequest(
			path,
			data,
			cfg.Langs,
			cfg.Options,
		),
	)

	if errors.Is(err, context.Canceled) {
		return guard.ValidateResponse{}, err
	}

	// Validation-level errors are represented in resp.
	return resp, nil
}

func (cfg fileRunConfig) markCanceledFiles(
	outcomes []fileOutcome,
	files []string,
	start int,
	err error,
) {
	for offset, path := range files {
		index := start + offset

		var b strings.Builder
		renderer := newFileRenderer(&b, cfg)

		renderer.fileHeader(index, path)

		outcomes[index] = cfg.opErrorOutcome(
			index,
			path,
			&b,
			err,
		)
	}
}
