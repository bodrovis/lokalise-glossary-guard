package guard

import (
	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/validator"
)

func newSummary(sum validator.Summary) Summary {
	outcomes := make([]Outcome, 0, len(sum.Outcomes))

	for _, o := range sum.Outcomes {
		outcomes = append(outcomes, Outcome{
			Name:     o.Result.Name,
			Status:   string(o.Result.Status),
			Message:  o.Result.Message,
			Critical: isCriticalCheck(o.Result.Name),
			Changed:  o.Final.DidChange,
			Note:     o.Final.Note,
		})
	}

	return Summary{
		FilePath:     sum.FilePath,
		Pass:         sum.Pass,
		Warn:         sum.Warn,
		Fail:         sum.Fail,
		Errors:       sum.Error,
		AppliedFixes: sum.AppliedFixes,
		EarlyExit:    sum.EarlyExit,
		EarlyCheck:   sum.EarlyCheck,
		EarlyStatus:  string(sum.EarlyStatus),
		FinalPath:    sum.FinalPath,
		Outcomes:     outcomes,
	}
}

func isCriticalCheck(name string) bool {
	cu, ok := checks.Lookup(name)
	return ok && cu.FailFast()
}
