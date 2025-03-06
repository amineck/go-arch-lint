package decoder

import (
	"github.com/amineck/go-arch-lint/internal/models/common"
)

type (
	yamlSourceCodeReferenceResolver interface {
		Resolve(filePath string, yamlPath string) common.Reference
	}

	jsonSchemaProvider interface {
		Provide(version int) ([]byte, error)
	}
)
