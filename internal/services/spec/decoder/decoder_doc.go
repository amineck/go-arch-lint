package decoder

import "github.com/amineck/go-arch-lint/internal/services/spec"

type doc interface {
	spec.Document

	postSetup()
}
