package assembler

import (
	"regexp"

	"github.com/amineck/go-arch-lint/internal/models/arch"
	"github.com/amineck/go-arch-lint/internal/models/common"
	"github.com/amineck/go-arch-lint/internal/services/spec"
)

type excludeFilesMatcherAssembler struct{}

func newExcludeFilesMatcherAssembler() *excludeFilesMatcherAssembler {
	return &excludeFilesMatcherAssembler{}
}

func (efa *excludeFilesMatcherAssembler) assemble(spec *arch.Spec, yamlSpec spec.Document) error {
	for _, regString := range yamlSpec.ExcludedFilesRegExp() {
		matcher, err := regexp.Compile(regString.Value)
		if err != nil {
			continue
		}

		spec.ExcludeFilesMatcher = append(spec.ExcludeFilesMatcher, common.NewReferable(
			matcher,
			regString.Reference,
		))
	}

	return nil
}
