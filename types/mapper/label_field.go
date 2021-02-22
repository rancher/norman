package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/values"
)

type LabelField struct {
	Field  string
	Domain string
}

func (e LabelField) FromInternal(data map[string]interface{}) {
	if len(e.Domain) == 0 {
		e.Domain = "field.cattle.io"
	}
	v, ok := values.RemoveValue(data, "labels", e.Domain+"/"+e.Field)
	if ok {
		data[e.Field] = v
	}
}

func (e LabelField) ToInternal(data map[string]interface{}) error {
	if len(e.Domain) == 0 {
		e.Domain = "field.cattle.io"
	}
	v, ok := data[e.Field]
	if ok {
		values.PutValue(data, v, "labels", e.Domain+"/"+e.Field)
	}
	return nil
}

func (e LabelField) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return ValidateField(e.Field, schema)
}
