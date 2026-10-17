// Throwaway probe: replicates BuildFlow's in-process SDK calls
// (detect via finding.NewDeviationDetector, repair via client.FixConfig)
// against this repo's own config, to prove detect/apply parity at the
// SDK boundary. Delete after the incident is resolved.
package main

import (
	"context"
	"fmt"
	"io"

	charmlog "charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/client"
	golaconfig "github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func main() {
	ctx := context.Background()
	const configPath = ".golangci.yml"

	logger := charmlog.New(io.Discard)
	analyzer := linter.NewAnalyzer(logger)
	loader := golaconfig.NewLoader(logger)
	planner := linter.NewFixer(logger, analyzer, loader)
	planner.SetGoVersionProvider(golaconfig.GetLocalGoVersion)

	changes, err := planner.PlanDeviations(ctx, configPath, types.LinterPriorityMedium)
	if err != nil {
		fmt.Println("PLAN ERROR:", err)

		return
	}

	fmt.Println("== PlanDeviations (Medium) ==")
	for _, c := range changes {
		fmt.Printf("  %-7s %-45s element=%q\n", c.Kind, c.Path, c.Element)
	}
	if len(changes) == 0 {
		fmt.Println("  (none)")
	}

	detector := finding.NewDeviationDetector(planner, configPath, types.LinterPriorityMedium)
	findings, err := detector.Detect(ctx)
	if err != nil {
		fmt.Println("DETECT ERROR:", err)

		return
	}

	fmt.Printf("== Detect findings: %d ==\n", len(findings))
	for _, f := range findings {
		fmt.Printf("  [%s] %s: %s\n", f.Severity, f.Rule, f.Message)
	}

	c := client.New(client.Options{})
	result, err := c.FixConfig(ctx, configPath, client.FixOptions{
		Priority: types.LinterPriorityMedium,
		DryRun:   true,
	})
	if err != nil {
		fmt.Println("FIXCONFIG ERROR:", err)

		return
	}

	fmt.Printf("== FixConfig dry-run: fixes=%d message=%q ==\n",
		result.FixesApplied, result.Message)

	plan, err := planner.PlanDeviations(ctx, configPath, types.LinterPriorityMedium)
	fmt.Printf("== Post-fix plan: %d changes (err=%v) ==\n", len(plan), err)
}
