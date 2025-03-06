package assembler

import (
	"github.com/amineck/go-arch-lint/internal/models/arch"
	"github.com/amineck/go-arch-lint/internal/services/spec"
)

type (
	archDecoder interface {
		Decode(archFile string) (spec.Document, []arch.Notice, error)
	}

	archValidator interface {
		Validate(doc spec.Document, projectDir string) []arch.Notice
	}

	pathResolver interface {
		Resolve(absPath string) (resolvePaths []string, err error)
	}
)
