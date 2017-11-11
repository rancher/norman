package mapper

import (
	"fmt"

	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
)

type Move struct {
	From, To string
}

func (m Move) FromInternal(data map[string]interface{}) {
	if v, ok := GetValue(data, m.From); ok {
		data[m.To] = v
	}
}

func (m Move) ToInternal(data map[string]interface{}) {
	if v, ok := GetValue(data, m.To); ok {
		data[m.From] = v
	}
}

func (m Move) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	internalSchema, err := getInternal(schema)
	if err != nil {
		return err
	}

	field, ok := internalSchema.ResourceFields[m.From]
	if !ok {
		return fmt.Errorf("missing field %s on internal schema %s", m.From, internalSchema.ID)
	}

	_, ok = schema.ResourceFields[m.To]
	if ok {
		return fmt.Errorf("field %s already exists on schema %s", m.From, internalSchema.ID)
	}

	delete(schema.ResourceFields, m.From)

	field.CodeName = convert.Capitalize(m.To)
	schema.ResourceFields[m.To] = field

	return nil
}
