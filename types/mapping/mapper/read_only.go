package mapper

import (
	"fmt"

	"github.com/rancher/norman/types"
)

type ReadOnly struct {
	Field string
}

func (r *ReadOnly) FromInternal(data map[string]interface{}) {
}

func (r *ReadOnly) ToInternal(data map[string]interface{}) {
}

func (r *ReadOnly) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	field, ok := schema.ResourceFields[r.Field]
	if !ok {
		return fmt.Errorf("failed to find field %s on schema %s", r.Field, schema.ID)
	}

	field.Create = false
	field.Update = false
	schema.ResourceFields[r.Field] = field

	return nil
}
