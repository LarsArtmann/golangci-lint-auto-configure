package finding

import (
	"context"

	finding "github.com/larsartmann/go-finding"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// DeviationPlanner is the narrowest interface the detector needs: the
// reconciler's report-only evaluation for one config. Decouples pkg/finding
// from pkg/linter's fixer plumbing.
type DeviationPlanner interface {
	PlanDeviations(ctx context.Context, configPath string, priority types.LinterPriority) ([]linter.Change, error)

	Evaluate(ctx context.Context, configPath string, priority types.LinterPriority) (*linter.Evaluation, error)
}

// DeviationDetector reports desired-state drift as findings. It runs the
// converge engine's report-only evaluation and converts it — lint mode and
// apply mode consume the identical evaluation, so detect/repair drift is
// structurally impossible.
type DeviationDetector struct {
	planner    DeviationPlanner
	configPath string
	priority   types.LinterPriority
}

// NewDeviationDetector creates a detector that reports config deviations for
// the given file at the given recommendation threshold.
func NewDeviationDetector(
	planner DeviationPlanner,
	configPath string,
	priority types.LinterPriority,
) *DeviationDetector {
	return &DeviationDetector{
		planner:    planner,
		configPath: configPath,
		priority:   priority,
	}
}

// Name returns the detector name.
func (d *DeviationDetector) Name() string {
	return "config-deviation"
}

// Detect evaluates the convergence (report-only) and converts the
// evaluation into findings: one per planned Change, plus the
// suppressed-re-enable and unsupported-recommendation findings a bare
// change list cannot carry (audit-only events, no config change behind
// them).
func (d *DeviationDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	evaluation, err := d.planner.Evaluate(ctx, d.configPath, d.priority)
	if err != nil {
		return nil, apperrors.WrapClassifiedf(err, "detector.evaluate",
			"evaluate config %s", d.configPath)
	}

	return EvaluationToFindings(evaluation)
}
