package validator

import (
	"github.com/amineck/go-arch-lint/internal/models/arch"
	"github.com/amineck/go-arch-lint/internal/services/spec"
)

type (
	validator interface {
		Validate(doc spec.Document) []arch.Notice
	}

	pathResolver interface {
		Resolve(absPath string) (resolvePaths []string, err error)
	}
)
