package assembler

import (
	"github.com/amineck/go-arch-lint/internal/models"
	"github.com/amineck/go-arch-lint/internal/models/common"
	"github.com/amineck/go-arch-lint/internal/services/spec"
)

type allowedVendorImportsAssembler struct {
	resolver *resolver
}

func newAllowedVendorImportsAssembler(
	resolver *resolver,
) *allowedVendorImportsAssembler {
	return &allowedVendorImportsAssembler{
		resolver: resolver,
	}
}

func (aia *allowedVendorImportsAssembler) assemble(
	yamlDocument spec.Document,
	vendorNames []string,
) ([]common.Referable[models.Glob], error) {
	list := make([]common.Referable[models.Glob], 0)

	allowedVendors := make([]string, 0)
	allowedVendors = append(allowedVendors, vendorNames...)
	for _, vendorName := range yamlDocument.CommonVendors() {
		allowedVendors = append(allowedVendors, vendorName.Value)
	}

	for _, name := range allowedVendors {
		yamlVendor, ok := yamlDocument.Vendors()[name]
		if !ok {
			continue
		}

		for _, vendorIn := range yamlVendor.Value.ImportPaths() {
			list = append(list, common.NewReferable(vendorIn, yamlVendor.Reference))
		}
	}

	return list, nil
}
