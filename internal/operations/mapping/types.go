package mapping

import (
	"context"

	"github.com/amineck/go-arch-lint/internal/models"
	"github.com/amineck/go-arch-lint/internal/models/arch"
	"github.com/amineck/go-arch-lint/internal/models/common"
)

type (
	specAssembler interface {
		Assemble(prj common.Project) (arch.Spec, error)
	}

	projectFilesResolver interface {
		ProjectFiles(ctx context.Context, spec arch.Spec) ([]models.FileHold, error)
	}

	projectInfoAssembler interface {
		ProjectInfo(rootDirectory string, archFilePath string) (common.Project, error)
	}
)
