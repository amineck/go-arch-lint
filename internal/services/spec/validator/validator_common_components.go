package validator

import (
	"github.com/amineck/go-arch-lint/internal/models/arch"
	"github.com/amineck/go-arch-lint/internal/services/spec"
)

type validatorCommonComponents struct {
	utils *utils
}

func newValidatorCommonComponents(
	utils *utils,
) *validatorCommonComponents {
	return &validatorCommonComponents{
		utils: utils,
	}
}

func (v *validatorCommonComponents) Validate(doc spec.Document) []arch.Notice {
	notices := make([]arch.Notice, 0)

	for _, componentName := range doc.CommonComponents() {
		if err := v.utils.assertKnownComponent(componentName.Value); err != nil {
			notices = append(notices, arch.Notice{
				Notice: err,
				Ref:    componentName.Reference,
			})
		}
	}

	return notices
}
