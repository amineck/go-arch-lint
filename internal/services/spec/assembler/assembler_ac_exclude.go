package assembler

import (
	"fmt"
	"path"

	"github.com/amineck/go-arch-lint/internal/models/arch"
	"github.com/amineck/go-arch-lint/internal/services/spec"
)

type excludeAssembler struct {
	resolver *resolver
}

func newExcludeAssembler(
	resolver *resolver,
) *excludeAssembler {
	return &excludeAssembler{
		resolver: resolver,
	}
}

func (ea *excludeAssembler) assemble(spec *arch.Spec, document spec.Document) error {
	for _, yamlRelativePath := range document.ExcludedDirectories() {
		tmpResolvedPath, err := ea.resolver.resolveLocalGlobPath(
			path.Clean(fmt.Sprintf("%s/%s",
				document.WorkingDirectory().Value,
				yamlRelativePath.Value,
			)),
		)
		if err != nil {
			return fmt.Errorf("failed to assemble exclude '%s' path's: %w", yamlRelativePath.Value, err)
		}

		resolvedPath := wrap(yamlRelativePath.Reference, tmpResolvedPath)
		spec.Exclude = append(spec.Exclude, resolvedPath...)
	}

	return nil
}
