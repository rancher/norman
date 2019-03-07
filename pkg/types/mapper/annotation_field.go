package mapper

import (
	"encoding/json"

	"github.com/rancher/norman/pkg/types"
	convert2 "github.com/rancher/norman/pkg/types/convert"
	values2 "github.com/rancher/norman/pkg/types/values"
)

type AnnotationField struct {
	Field            string
	Object           bool
	List             bool
	IgnoreDefinition bool
}

func (e AnnotationField) FromInternal(data map[string]interface{}) {
	v, ok := values2.RemoveValue(data, "annotations", "field.cattle.io/"+e.Field)
	if ok {
		if e.Object {
			data := map[string]interface{}{}
			//ignore error
			if err := json.Unmarshal([]byte(convert2.ToString(v)), &data); err == nil {
				v = data
			}
		}
		if e.List {
			var data []interface{}
			if err := json.Unmarshal([]byte(convert2.ToString(v)), &data); err == nil {
				v = data
			}
		}

		data[e.Field] = v
	}
}

func (e AnnotationField) ToInternal(data map[string]interface{}) error {
	v, ok := data[e.Field]
	if ok {
		if e.Object || e.List {
			if bytes, err := json.Marshal(v); err == nil {
				v = string(bytes)
			}
		}
		values2.PutValue(data, convert2.ToString(v), "annotations", "field.cattle.io/"+e.Field)
	}
	values2.RemoveValue(data, e.Field)
	return nil
}

func (e AnnotationField) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	if e.IgnoreDefinition {
		return nil
	}
	return ValidateField(e.Field, schema)
}
