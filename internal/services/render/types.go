package render

import (
	"github.com/amineck/go-arch-lint/internal/models/common"
)

type (
	referenceRender interface {
		SourceCode(ref common.Reference, highlight bool, showPointer bool) []byte
	}

	colorPrinter interface {
		Red(in string) (out string)
		Green(in string) (out string)
		Yellow(in string) (out string)
		Blue(in string) (out string)
		Magenta(in string) (out string)
		Cyan(in string) (out string)
		Gray(in string) (out string)
	}
)
