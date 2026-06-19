package validate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

type job struct {
	idx  int
	path string
}

type validateRunConfig struct {
	Files       []string
	Langs       []string
	MaxParallel uint
	Options     checks.RunOptions
	JSONOut     bool
}

func runValidate(ctx context.Context, cfg validateRunConfig) error {
	start := time.Now()
	sep := strings.Repeat("─", 72)

	outcomes := runFiles(ctx, cfg.Files, cfg.Langs, sep, cfg.MaxParallel, cfg.Options)

	return finalize(outcomes, len(cfg.Files), start, cfg.JSONOut)
}

func runFiles(
	ctx context.Context,
	files []string,
	langs []string,
	sep string,
	maxParallel uint,
	opts checks.RunOptions,
) []fileOutcome {
	jobs := make(chan job)
	outcomes := make([]fileOutcome, len(files))

	workers := workerCount(maxParallel, len(files))

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()

			for j := range jobs {
				outcomes[j.idx] = runOneFile(ctx, j.idx, j.path, langs, sep, opts)
			}
		}()
	}

	go func() {
		defer close(jobs)

		for i, p := range files {
			select {
			case <-ctx.Done():
				for j := i; j < len(files); j++ {
					var b strings.Builder
					renderFileHeader(&b, j, files[j], sep, opts)
					outcomes[j] = fileOpErrorOutcome(j, files[j], &b, ctx.Err(), sep)
				}
				return
			case jobs <- job{idx: i, path: p}:
			}
		}
	}()

	wg.Wait()

	return outcomes
}

func workerCount(maxParallel uint, filesCount int) int {
	if maxParallel < 1 {
		maxParallel = uint(runtime.GOMAXPROCS(0))
	}

	workers := min(int(maxParallel), filesCount)

	return max(1, workers)
}

func runOneFile(ctx context.Context, i int, path string, langs []string, sep string, opts checks.RunOptions) fileOutcome {
	var b strings.Builder

	renderFileHeader(&b, i, path, sep, opts)

	data, err := os.ReadFile(path)
	if err != nil {
		return fileOpErrorOutcome(i, path, &b, err, sep)
	}

	resp, validationErr := guard.ValidateBytes(ctx, buildValidateRequest(path, data, langs, opts))
	if errors.Is(validationErr, context.Canceled) {
		return fileOpErrorOutcome(i, path, &b, validationErr, sep)
	}

	if validationErr != nil {
		// Validation-level errors are represented in resp.
		// Keep old CLI behavior: don't print an extra raw Go error here.
		_ = validationErr
	}

	oc := fileOutcome{Idx: i, Path: path}
	applyGuardResponse(&oc, resp)

	renderValidationReport(&b, path, resp.Summary)

	if writeErr := writeFixedFileIfNeeded(&b, opts, resp); writeErr != nil {
		oc.HadOpErr = true
		oc.Errored++
	}

	renderResult(&b, resp)

	fmt.Fprintf(&b, "%s\n", sep)
	oc.Output = b.String()

	return oc
}
