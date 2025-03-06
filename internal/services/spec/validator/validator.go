package validator

import (
	"github.com/amineck/go-arch-lint/internal/models/arch"
	"github.com/amineck/go-arch-lint/internal/services/spec"
)

type Validator struct {
	pathResolver pathResolver
}

func NewValidator(
	pathResolver pathResolver,
) *Validator {
	return &Validator{
		pathResolver: pathResolver,
	}
}

func (v *Validator) Validate(doc spec.Document, projectDir string) []arch.Notice {
	notices := make([]arch.Notice, 0)

	utils := newUtils(v.pathResolver, doc, projectDir)
	validators := []validator{
		newValidatorCommonComponents(utils),
		newValidatorCommonVendors(utils),
		newValidatorComponents(utils),
		newValidatorDeps(utils),
		newValidatorDepsComponents(utils),
		newValidatorDepsVendors(utils),
		newValidatorExcludeFiles(),
		newValidatorVendors(utils),
		newValidatorVersion(),
		newValidatorWorkDir(utils),
	}

	for _, specValidator := range validators {
		notices = append(notices, specValidator.Validate(doc)...)
	}

	return notices
}
