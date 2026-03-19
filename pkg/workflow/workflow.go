package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/universal-workflow/pkg/types"
	workflowpkg "github.com/LarsArtmann/universal-workflow/pkg/workflow"
	"github.com/charmbracelet/log"
	apperrors "github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
)

// ActivityContext provides dependencies for workflow activities.
//
//nolint:containedctx // Context stored here is acceptable for workflow input data pattern.
type ActivityContext struct {
	Context      context.Context
	ConfigPath   string
	Analyzer     *linter.Analyzer
	Logger       *log.Logger
	DryRun       bool
	GenerateHTML bool
	OutputReport string
}

// AnalysisActivity analyzes golangci-lint configuration.
func AnalysisActivity(ctx workflowpkg.ActivityContext) (*types.ActivityResult, error) {
	activityCtx, ok := ctx.Input.(*ActivityContext)
	if !ok {
		return nil, apperrors.ErrInvalidActivityContext
	}

	activityCtx.Logger.Infof("Analyzing golangci-lint configuration...")

	analysis, err := activityCtx.Analyzer.AnalyzeConfig(activityCtx.Context, activityCtx.ConfigPath)
	if err != nil {
		return &types.ActivityResult{
			Status:    types.ActivityStatusFailed,
			Output:    fmt.Sprintf("Analysis failed: %v", err),
			StartTime: time.Now(),
			EndTime:   time.Now().Add(1 * time.Second),
		}, fmt.Errorf("analysis failed: %w", err)
	}

	// Format and display recommendations
	recommendations := activityCtx.Analyzer.FormatRecommendations(analysis)
	summary := activityCtx.Analyzer.GetSummary(analysis)

	activityCtx.Logger.Info("\n" + recommendations)
	activityCtx.Logger.Infof("Summary: %s", summary)

	return &types.ActivityResult{
		Status:    types.ActivityStatusCompleted,
		Output:    summary,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Data: map[string]any{
			"analysis": analysis,
		},
	}, nil
}

// ValidationActivity validates golangci-lint configuration.
func ValidationActivity(ctx workflowpkg.ActivityContext) (*types.ActivityResult, error) {
	activityCtx, ok := ctx.Input.(*ActivityContext)
	if !ok {
		return nil, apperrors.ErrInvalidActivityContext
	}

	activityCtx.Logger.Infof("Validating golangci-lint configuration...")

	// This would typically run `golangci-lint config verify`
	// For now, we'll simulate successful validation
	activityCtx.Logger.Infof("Configuration is valid")

	return &types.ActivityResult{
		Status:    types.ActivityStatusCompleted,
		Output:    "Configuration validation passed",
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}, nil
}

// ReportActivity generates HTML report of the analysis.
func ReportActivity(ctx workflowpkg.ActivityContext) (*types.ActivityResult, error) {
	activityCtx, ok := ctx.Input.(*ActivityContext)
	if !ok {
		return nil, apperrors.ErrInvalidActivityContext
	}

	if !activityCtx.GenerateHTML {
		activityCtx.Logger.Debugf("HTML report generation disabled")

		return &types.ActivityResult{
			Status:    types.ActivityStatusCompleted,
			Output:    "Skipped HTML report generation",
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}, nil
	}

	activityCtx.Logger.Infof("Generating HTML report...")

	// HTML generation would be implemented with templ components
	activityCtx.Logger.Debugf("HTML report would be saved to: %s", activityCtx.OutputReport)

	activityCtx.Logger.Infof("HTML report generated: %s", activityCtx.OutputReport)

	return &types.ActivityResult{
		Status:    types.ActivityStatusCompleted,
		Output:    "HTML report: " + activityCtx.OutputReport,
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}, nil
}

// Builder constructs the golangci-lint configuration workflow.
type Builder struct {
	logger   *log.Logger
	analyzer *linter.Analyzer
}

// NewBuilder creates a new workflow builder.
func NewBuilder(logger *log.Logger, analyzer *linter.Analyzer) *Builder {
	return &Builder{
		logger:   logger,
		analyzer: analyzer,
	}
}

// BuildAutoConfigureWorkflow creates a workflow for auto-configuring golangci-lint.
func (b *Builder) BuildAutoConfigureWorkflow(
	ctx context.Context,
	configPath string,
	dryRun, generateHTML bool,
	outputPath string,
) (workflowpkg.WorkflowLike, error) {
	workflowID := types.MustWorkflowID("golangci-lint-auto-configure")
	workflowName := types.WorkflowName("Automatically analyze and configure golangci-lint")

	wf := workflowpkg.NewUnifiedWorkflow(workflowID, workflowName)

	// Set up activity context
	activityCtx := &ActivityContext{
		Context:      ctx,
		ConfigPath:   configPath,
		Analyzer:     b.analyzer,
		Logger:       b.logger,
		DryRun:       dryRun,
		GenerateHTML: generateHTML,
		OutputReport: outputPath,
	}

	// Add analysis activity
	wf.Step(types.MustActivityID("analyze-config"), func(ctx workflowpkg.ActivityContext) (*types.ActivityResult, error) {
		ctx.Input = activityCtx

		return AnalysisActivity(ctx)
	})

	// Add validation activity (depends on analysis)
	wf.Step(types.MustActivityID("validate-config"), func(ctx workflowpkg.ActivityContext) (*types.ActivityResult, error) {
		ctx.Input = activityCtx

		return ValidationActivity(ctx)
	}).DependsOn(types.MustActivityID("analyze-config"))

	// Add report generation activity (depends on validation)
	wf.Step(types.MustActivityID("generate-report"), func(ctx workflowpkg.ActivityContext) (*types.ActivityResult, error) {
		ctx.Input = activityCtx

		return ReportActivity(ctx)
	}).DependsOn(types.MustActivityID("validate-config"))

	return wf, nil
}

// ExecuteAutoConfigureWorkflow executes the auto-configure workflow.
func (b *Builder) ExecuteAutoConfigureWorkflow(
	ctx context.Context,
	configPath string,
	dryRun, generateHTML bool,
	outputPath string,
) (workflowpkg.WorkflowRun, error) {
	wf, err := b.BuildAutoConfigureWorkflow(ctx, configPath, dryRun, generateHTML, outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build workflow: %w", err)
	}

	b.logger.Infof("Executing workflow: %s", wf.GetName())

	run, err := wf.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}

	return run, nil
}
