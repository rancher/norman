package mapper

import (
	"fmt"

	"github.com/rancher/norman/types"
)

func getInternal(schema *types.Schema) (*types.Schema, error) {
	if schema.InternalSchema == nil {
		return nil, fmt.Errorf("no internal schema found for schema %s", schema.ID)
	}

	return schema.InternalSchema, nil
}

func validateInternalField(field string, schema *types.Schema) (*types.Schema, error) {
	internalSchema, err := getInternal(schema)
	if err != nil {
		return nil, err
	}

	if _, ok := internalSchema.ResourceFields[field]; !ok {
		return nil, fmt.Errorf("field %s missing on internal schema %s", field, schema.ID)
	}

	return internalSchema, nil
}
