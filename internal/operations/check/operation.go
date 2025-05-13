package check

import (
	"context"
	"fmt"
	"sort"

	"github.com/fe3dback/go-arch-lint/internal/models"
	"github.com/fe3dback/go-arch-lint/internal/models/arch"
)

type (
	Operation struct {
		projectInfoAssembler projectInfoAssembler
		specAssembler        specAssembler
		specChecker          specChecker
		referenceRender      referenceRender
		highlightCodePreview bool
	}

	limiterResult struct {
		results      models.CheckResult
		omittedCount int
	}
)

func NewOperation(
	projectInfoAssembler projectInfoAssembler,
	specAssembler specAssembler,
	specChecker specChecker,
	referenceRender referenceRender,
	highlightCodePreview bool,
) *Operation {
	return &Operation{
		projectInfoAssembler: projectInfoAssembler,
		specAssembler:        specAssembler,
		specChecker:          specChecker,
		referenceRender:      referenceRender,
		highlightCodePreview: highlightCodePreview,
	}
}

func (o *Operation) Behave(ctx context.Context, in models.CmdCheckIn) (models.CmdCheckOut, error) {
	projectInfo, err := o.projectInfoAssembler.ProjectInfo(in.ProjectPath, in.ArchFile)
	if err != nil {
		return models.CmdCheckOut{}, fmt.Errorf("failed to assemble project info: %w", err)
	}

	spec, err := o.specAssembler.Assemble(projectInfo)
	if err != nil {
		return models.CmdCheckOut{}, fmt.Errorf("failed to assemble spec: %w", err)
	}

	result := models.CheckResult{}
	if len(spec.Integrity.DocumentNotices) == 0 {
		result, err = o.specChecker.Check(ctx, spec)
		if err != nil {
			return models.CmdCheckOut{}, fmt.Errorf("failed to check project deps: %w", err)
		}
	}

	limitedResult := o.limitResults(result, in.MaxWarnings)

	model := models.CmdCheckOut{
		ModuleName:             spec.ModuleName.Value,
		DocumentNotices:        o.assembleNotice(spec.Integrity),
		ArchHasWarnings:        o.resultsHasWarnings(limitedResult.results),
		ArchWarningsDependency: limitedResult.results.DependencyWarnings,
		ArchWarningsMatch:      limitedResult.results.MatchWarnings,
		ArchWarningsDeepScan:   limitedResult.results.DeepscanWarnings,
		OmittedCount:           limitedResult.omittedCount,
		Qualities: []models.CheckQuality{
			{
				ID:   "component_imports",
				Name: "Base: component imports",
				Used: len(spec.Components) > 0,
				Hint: "always on",
			},
			{
				ID:   "vendor_imports",
				Name: "Advanced: vendor imports",
				Used: spec.Allow.DepOnAnyVendor.Value == false,
				Hint: "switch 'allow.depOnAnyVendor = false' (or delete) to on",
			},
			{
				ID:   "deepscan",
				Name: "Advanced: method calls and dependency injections",
				Used: spec.Allow.DeepScan.Value == true,
				Hint: "switch 'allow.deepScan = true' (or delete) to on",
			},
		},
	}

	if model.ArchHasWarnings || len(model.DocumentNotices) > 0 {
		// normal output with exit code 1
		return model, models.NewUserSpaceError("check not successful")
	}

	return model, nil
}

func (o *Operation) limitResults(result models.CheckResult, maxWarnings int) limiterResult {
	passCount := 0
	limitedResults := models.CheckResult{
		DependencyWarnings: []models.CheckArchWarningDependency{},
		MatchWarnings:      []models.CheckArchWarningMatch{},
		DeepscanWarnings:   []models.CheckArchWarningDeepscan{},
	}

	// append deps
	for _, notice := range result.DependencyWarnings {
		if passCount >= maxWarnings {
			break
		}

		limitedResults.DependencyWarnings = append(limitedResults.DependencyWarnings, notice)
		passCount++
	}

	// append not matched
	for _, notice := range result.MatchWarnings {
		if passCount >= maxWarnings {
			break
		}

		limitedResults.MatchWarnings = append(limitedResults.MatchWarnings, notice)
		passCount++
	}

	// append deep scan
	for _, notice := range result.DeepscanWarnings {
		if passCount >= maxWarnings {
			break
		}

		limitedResults.DeepscanWarnings = append(limitedResults.DeepscanWarnings, notice)
		passCount++
	}

	totalCount := 0 +
		len(result.DeepscanWarnings) +
		len(result.DependencyWarnings) +
		len(result.MatchWarnings)

	return limiterResult{
		results:      limitedResults,
		omittedCount: totalCount - passCount,
	}
}

func (o *Operation) resultsHasWarnings(result models.CheckResult) bool {
	if len(result.DependencyWarnings) > 0 {
		return true
	}

	if len(result.MatchWarnings) > 0 {
		return true
	}

	if len(result.DeepscanWarnings) > 0 {
		return true
	}

	return false
}

func (o *Operation) assembleNotice(integrity arch.Integrity) []models.CheckNotice {
	notices := make([]arch.Notice, 0)
	notices = append(notices, integrity.DocumentNotices...)

	results := make([]models.CheckNotice, 0)
	for _, notice := range notices {
		results = append(results, models.CheckNotice{
			Text:   fmt.Sprintf("%s", notice.Notice),
			File:   notice.Ref.File,
			Line:   notice.Ref.Line,
			Column: notice.Ref.Column,
			SourceCodePreview: o.referenceRender.SourceCode(
				notice.Ref.ExtendRange(1, 1),
				o.highlightCodePreview,
				true,
			),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		sI := results[i]
		sJ := results[j]

		if sI.File == sJ.File {
			return sI.Line < sJ.Line
		}

		return sI.File < sJ.File
	})

	return results
}
