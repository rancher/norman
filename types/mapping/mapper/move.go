package mapper

import (
	"fmt"

	"strings"

	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
)

type Move struct {
	From, To string
}

func (m Move) FromInternal(data map[string]interface{}) {
	if v, ok := RemoveValue(data, strings.Split(m.From, "/")...); ok {
		PutValue(data, v, strings.Split(m.To, "/")...)
	}
}

func (m Move) ToInternal(data map[string]interface{}) {
	if v, ok := RemoveValue(data, strings.Split(m.To, "/")...); ok {
		PutValue(data, v, strings.Split(m.From, "/")...)
	}
}

func (m Move) ModifySchema(s *types.Schema, schemas *types.Schemas) error {
	internalSchema, err := getInternal(s)
	if err != nil {
		return err
	}

	_, _, fromInternalField, ok, err := getField(internalSchema, schemas, m.From)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("missing field %s on internal schema %s", m.From, internalSchema.ID)
	}

	fromSchema, _, _, _, err := getField(s, schemas, m.From)
	if err != nil {
		return err
	}

	toSchema, toFieldName, toField, ok, err := getField(s, schemas, m.To)
	if err != nil {
		return err
	}
	_, ok = toSchema.ResourceFields[toFieldName]
	if ok && !strings.Contains(m.To, "/") {
		return fmt.Errorf("field %s already exists on schema %s", m.To, s.ID)
	}

	delete(fromSchema.ResourceFields, m.From)

	toField.CodeName = convert.Capitalize(toFieldName)
	toSchema.ResourceFields[toFieldName] = fromInternalField

	return nil
}

func getField(schema *types.Schema, schemas *types.Schemas, target string) (*types.Schema, string, types.Field, bool, error) {
	parts := strings.Split(target, "/")
	for i, part := range parts {
		if i == len(parts)-1 {
			continue
		}

		subSchema := schemas.Schema(&schema.Version, schema.ResourceFields[part].Type)
		if subSchema == nil {
			return nil, "", types.Field{}, false, fmt.Errorf("failed to find field or schema for %s on %s", part, schema.ID)
		}

		schema = subSchema
	}

	name := parts[len(parts)-1]
	f, ok := schema.ResourceFields[name]
	return schema, name, f, ok, nil
}
