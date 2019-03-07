package mapper

import (
	"github.com/rancher/norman/pkg/types"
	convert2 "github.com/rancher/norman/pkg/types/convert"
)

type SliceMerge struct {
	From             []string
	To               string
	IgnoreDefinition bool
}

func (s SliceMerge) FromInternal(data map[string]interface{}) {
	var result []interface{}
	for _, name := range s.From {
		val, ok := data[name]
		if !ok {
			continue
		}
		result = append(result, convert2.ToInterfaceSlice(val)...)
	}

	if result != nil {
		data[s.To] = result
	}
}

func (s SliceMerge) ToInternal(data map[string]interface{}) error {
	return nil
}

func (s SliceMerge) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	if s.IgnoreDefinition {
		return nil
	}

	for _, from := range s.From {
		if err := ValidateField(from, schema); err != nil {
			return err
		}
		if from != s.To {
			delete(schema.ResourceFields, from)
		}
	}

	return ValidateField(s.To, schema)
}
