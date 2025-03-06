package graph

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/amineck/go-arch-lint/internal/models"
	"github.com/amineck/go-arch-lint/internal/models/arch"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/textmeasure"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
)

type Operation struct {
	specAssembler        specAssembler
	projectInfoAssembler projectInfoAssembler
}

func NewOperation(
	specAssembler specAssembler,
	projectInfoAssembler projectInfoAssembler,
) *Operation {
	return &Operation{
		specAssembler:        specAssembler,
		projectInfoAssembler: projectInfoAssembler,
	}
}

func (o *Operation) Behave(ctx context.Context, in models.CmdGraphIn) (models.CmdGraphOut, error) {
	projectInfo, err := o.projectInfoAssembler.ProjectInfo(in.ProjectPath, in.ArchFile)
	if err != nil {
		return models.CmdGraphOut{}, fmt.Errorf("failed to assemble project info: %w", err)
	}

	spec, err := o.specAssembler.Assemble(projectInfo)
	if err != nil {
		return models.CmdGraphOut{}, fmt.Errorf("failed to assemble spec: %w", err)
	}

	graphCode, err := o.buildGraph(spec, in)
	if err != nil {
		return models.CmdGraphOut{}, fmt.Errorf("failed build graph: %w", err)
	}

	svg, err := o.compileGraph(ctx, graphCode)
	if err != nil {
		return models.CmdGraphOut{}, fmt.Errorf("failed to compile graph: %w", err)
	}

	outFile, err := filepath.Abs(in.OutFile)
	if err != nil {
		return models.CmdGraphOut{}, fmt.Errorf("failed get abs path from '%s': %w", in.OutFile, err)
	}

	if o.isFileShouldBeWritten(in) {
		err = os.WriteFile(outFile, svg, os.ModePerm)
		if err != nil {
			return models.CmdGraphOut{}, fmt.Errorf("failed write graph into '%s' file: %w", in.OutFile, err)
		}
	}

	return models.CmdGraphOut{
		ProjectDirectory: spec.RootDirectory.Value,
		ModuleName:       spec.ModuleName.Value,
		OutFile:          outFile,
		D2Definitions:    string(graphCode),
		ExportD2:         in.ExportD2,
	}, nil
}

func (o *Operation) isFileShouldBeWritten(in models.CmdGraphIn) bool {
	if in.OutputType == models.OutputTypeJSON {
		return false
	}

	if in.ExportD2 {
		return false
	}

	return true
}

func (o *Operation) buildGraph(spec arch.Spec, opts models.CmdGraphIn) ([]byte, error) {
	whiteList, err := o.populateGraphWhitelist(spec, opts)
	if err != nil {
		return nil, err
	}

	flow := o.componentsFlowArrow(opts)

	linesBuff := make([]string, 0, 256)

	for _, cmp := range spec.Components {
		if _, visible := whiteList[cmp.Name.Value]; !visible {
			continue
		}

		for _, dep := range cmp.MayDependOn {
			if _, visible := whiteList[dep.Value]; !visible {
				continue
			}

			linesBuff = append(linesBuff, fmt.Sprintf("%s %s %s\n", cmp.Name.Value, flow, dep.Value))
		}

		if opts.IncludeVendors {
			for _, vnd := range cmp.CanUse {
				vars := map[string]string{
					"vnd": vnd.Value,
					"cmp": cmp.Name.Value,
				}

				tpl := `
				{{vnd}}.style.font-size: 12
				{{vnd}}.style.stroke: "#77AA44"
				{{cmp}} <- {{vnd}} {
				  style.stroke: "#77AA44"
				  source-arrowhead: {
				    shape: diamond
				    style.filled: false
				  }
				}
				`

				for name, value := range vars {
					tpl = strings.ReplaceAll(tpl, fmt.Sprintf("{{%s}}", name), value)
				}
				linesBuff = append(linesBuff, tpl)
			}
		}
	}

	var buff bytes.Buffer
	sort.Strings(linesBuff)

	for _, line := range linesBuff {
		buff.WriteString(strings.ReplaceAll(line, "\t", ""))
	}

	return buff.Bytes(), nil
}

func (o *Operation) componentsFlowArrow(opts models.CmdGraphIn) string {
	if opts.Type == models.GraphTypeFlow {
		return "->"
	}

	if opts.Type == models.GraphTypeDI {
		return "<-"
	}

	return "--"
}

func (o *Operation) populateGraphWhitelist(spec arch.Spec, opts models.CmdGraphIn) (map[string]struct{}, error) {
	if opts.Focus == "" {
		return o.populateGraphWhitelistAll(spec)
	}

	return o.populateGraphWhitelistFocused(spec, opts.Focus)
}

func (o *Operation) populateGraphWhitelistAll(spec arch.Spec) (map[string]struct{}, error) {
	whiteList := make(map[string]struct{}, len(spec.Components))

	for _, cmp := range spec.Components {
		whiteList[cmp.Name.Value] = struct{}{}
	}

	return whiteList, nil
}

func (o *Operation) populateGraphWhitelistFocused(spec arch.Spec, focusCmpName string) (map[string]struct{}, error) {
	cmpMap := make(map[string]arch.Component)
	rootExist := false

	for _, cmp := range spec.Components {
		cmpMap[cmp.Name.Value] = cmp

		if focusCmpName == cmp.Name.Value {
			rootExist = true
		}
	}

	if !rootExist {
		return nil, fmt.Errorf("focused cmp %s is not defined", focusCmpName)
	}

	whiteList := make(map[string]struct{}, len(spec.Components))
	resolved := make(map[string]struct{}, 64)
	resolveList := make([]string, 0, 64)
	resolveList = append(resolveList, focusCmpName)

	for len(resolveList) > 0 {
		cmp := cmpMap[resolveList[0]]
		resolveList = resolveList[1:]

		if _, alreadyResolved := resolved[cmp.Name.Value]; alreadyResolved {
			continue
		}

		// cmp itself
		whiteList[cmp.Name.Value] = struct{}{}

		// cmp deps
		for _, dep := range cmp.MayDependOn {
			whiteList[dep.Value] = struct{}{}
			resolveList = append(resolveList, dep.Value)
		}

		// mark as resolved (for recursion check)
		resolved[cmp.Name.Value] = struct{}{}
	}

	return whiteList, nil
}

func (o *Operation) compileGraph(ctx context.Context, graphCode []byte) ([]byte, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("failed create ruler: %w", err)
	}

	diagram, _, err := d2lib.Compile(ctx, string(graphCode), &d2lib.CompileOptions{
		Layout: func(ctx context.Context, g *d2graph.Graph) error {
			return d2dagrelayout.Layout(ctx, g, nil)
		},
		Ruler: ruler,
	})
	if err != nil {
		return nil, fmt.Errorf("failed compile d2 graph: %w", err)
	}

	out, err := d2svg.Render(diagram, &d2svg.RenderOpts{
		Pad:     d2svg.DEFAULT_PADDING,
		Sketch:  true,
		ThemeID: d2themescatalog.NeutralDefault.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("svg render failed: %w", err)
	}

	return out, nil
}
