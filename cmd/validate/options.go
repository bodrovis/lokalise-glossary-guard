package validate

import (
	"runtime"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

type commandOptions struct {
	files       []string
	langs       []string
	maxParallel uint
	jsonOut     bool
	noColor     bool

	doFix         bool
	hardFailOnErr bool
	rerunAfterFix bool
}

func defaultCommandOptions() commandOptions {
	return commandOptions{
		maxParallel:   uint(runtime.GOMAXPROCS(0)),
		rerunAfterFix: true,
	}
}

func buildRunOptions(opts commandOptions) checks.RunOptions {
	fm := checks.FixNone
	if opts.doFix {
		fm = checks.FixIfNotPass
	}

	return checks.RunOptions{
		FixMode:       fm,
		RerunAfterFix: opts.rerunAfterFix,
		HardFailOnErr: opts.hardFailOnErr,
	}
}
