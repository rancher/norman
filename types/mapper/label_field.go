package mapper

import "github.com/rancher/norman/types"

type LabelField struct {
	Field string
}

func (e LabelField) FromInternal(data map[string]interface{}) {
	v, ok := RemoveValue(data, "labels", "io.cattle.field."+e.Field)
	if ok {
		data[e.Field] = v
	}
}

func (e LabelField) ToInternal(data map[string]interface{}) {
	v, ok := data[e.Field]
	if ok {
		PutValue(data, v, "labels", "io.cattle.field."+e.Field)
	}
}

func (e LabelField) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return validateField(e.Field, schema)
}
