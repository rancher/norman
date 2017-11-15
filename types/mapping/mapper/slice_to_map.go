package mapper

import (
	"fmt"

	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/definition"
)

type SliceToMap struct {
	Field string
	Key   string
}

func (s SliceToMap) FromInternal(data map[string]interface{}) {
	datas, _ := data[s.Field].([]interface{})
	result := map[string]interface{}{}

	for _, item := range datas {
		if mapItem, ok := item.(map[string]interface{}); ok {
			name, _ := mapItem[s.Key].(string)
			delete(mapItem, s.Key)
			result[name] = mapItem
		}
	}

	if len(result) > 0 {
		data[s.Field] = result
	}
}

func (s SliceToMap) ToInternal(data map[string]interface{}) {
	datas, _ := data[s.Field].(map[string]interface{})
	result := []map[string]interface{}{}

	for name, item := range datas {
		mapItem, _ := item.(map[string]interface{})
		if mapItem != nil {
			mapItem[s.Key] = name
			result = append(result, mapItem)
		}
	}

	if len(result) > 0 {
		data[s.Field] = result
	}
}

func (s SliceToMap) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	internalSchema, err := validateInternalField(s.Field, schema)
	if err != nil {
		return err
	}

	field := internalSchema.ResourceFields[s.Field]
	if !definition.IsArrayType(field.Type) {
		return fmt.Errorf("field %s on %s is not an array", s.Field, internalSchema.ID)
	}

	field.Type = "map[" + definition.SubType(field.Type) + "]"
	schema.ResourceFields[s.Field] = field

	return nil
}
