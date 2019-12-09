package mapper

import (
	"fmt"

	"github.com/rancher/norman/v2/pkg/types"
)

func ValidateField(field string, schema *types.Schema) error {
	if _, ok := schema.ResourceFields[field]; !ok {
		return fmt.Errorf("field %s missing on schema %s", field, schema.ID)
	}

	return nil
}
