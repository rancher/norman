package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/registry"
	"github.com/rancher/norman/store"
)

var (
	Create = Operation("create")
	Update = Operation("update")
	Action = Operation("action")
)

type Operation string

type Builder struct {
	Registry     registry.SchemaRegistry
	RefValidator store.ReferenceValidator
}

func (b *Builder) Construct(schema *client.Schema, input map[string]interface{}, op Operation) (map[string]interface{}, error) {
	return b.copyFields(schema, input, op)
}

func (b *Builder) copyInputs(schema *client.Schema, input map[string]interface{}, op Operation, result map[string]interface{}) error {
	for fieldName, value := range input {
		field, ok := schema.ResourceFields[fieldName]
		if !ok {
			continue
		}

		if !fieldMatchesOp(field, op) {
			continue
		}

		wasNull := value == nil && (field.Nullable || field.Default == nil)
		value, err := b.convert(field.Type, value, op)
		if err != nil {
			return httperror.WrapFieldAPIError(err, httperror.INVALID_FORMAT, fieldName, err.Error())
		}

		if value != nil || wasNull {
			if slice, ok := value.([]interface{}); ok {
				for _, sliceValue := range slice {
					if sliceValue == nil {
						return httperror.NewFieldAPIError(httperror.NOT_NULLABLE, fieldName, "Individual array values can not be null")
					}
					if err := checkFieldCriteria(fieldName, field, sliceValue); err != nil {
						return err
					}
				}
			} else {
				if err := checkFieldCriteria(fieldName, field, value); err != nil {
					return err
				}
			}
			result[fieldName] = value
		}
	}

	return nil
}

func (b *Builder) checkDefaultAndRequired(schema *client.Schema, input map[string]interface{}, op Operation, result map[string]interface{}) error {
	for fieldName, field := range schema.ResourceFields {
		_, hasKey := result[fieldName]
		if op == Create && !hasKey && field.Default != nil {
			result[fieldName] = field.Default
		}

		_, hasKey = result[fieldName]
		if op == Create && fieldMatchesOp(field, Create) && field.Required {
			if !hasKey {
				return httperror.NewFieldAPIError(httperror.MISSING_REQUIRED, fieldName, "")
			}

			if isArrayType(field.Type) {
				slice, err := b.convertArray(fieldName, result[fieldName], op)
				if err != nil {
					return err
				}
				if len(slice) == 0 {
					return httperror.NewFieldAPIError(httperror.MISSING_REQUIRED, fieldName, "")
				}
			}
		}
	}

	return nil
}

func (b *Builder) copyFields(schema *client.Schema, input map[string]interface{}, op Operation) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	if err := b.copyInputs(schema, input, op, result); err != nil {
		return nil, err
	}

	return result, b.checkDefaultAndRequired(schema, input, op, result)
}

func checkFieldCriteria(fieldName string, field client.Field, value interface{}) error {
	numVal, isNum := value.(int64)
	strVal := ""
	hasStrVal := false

	if value == nil && field.Default != nil {
		value = field.Default
	}

	if value != nil {
		hasStrVal = true
		strVal = fmt.Sprint(value)
	}

	if value == nil && !field.Nullable {
		return httperror.NewFieldAPIError(httperror.NOT_NULLABLE, fieldName, "")
	}

	if isNum {
		if field.Min != nil && numVal < *field.Min {
			return httperror.NewFieldAPIError(httperror.MIN_LIMIT_EXCEEDED, fieldName, "")
		}
		if field.Max != nil && numVal > *field.Max {
			return httperror.NewFieldAPIError(httperror.MAX_LIMIT_EXCEEDED, fieldName, "")
		}
	}

	if hasStrVal {
		if field.MinLength != nil && len(strVal) < *field.MinLength {
			return httperror.NewFieldAPIError(httperror.MIN_LENGTH_EXCEEDED, fieldName, "")
		}
		if field.MaxLength != nil && len(strVal) > *field.MaxLength {
			return httperror.NewFieldAPIError(httperror.MAX_LENGTH_EXCEEDED, fieldName, "")
		}
	}

	if len(field.Options) > 0 {
		if hasStrVal || !field.Nullable {
			found := false
			for _, option := range field.Options {
				if strVal == option {
					found = true
					break
				}
			}

			if !found {
				httperror.NewFieldAPIError(httperror.INVALID_OPTION, fieldName, "")
			}
		}
	}

	if len(field.ValidChars) > 0 && hasStrVal {
		for _, c := range strVal {
			if !strings.ContainsRune(field.ValidChars, c) {
				httperror.NewFieldAPIError(httperror.INVALID_CHARACTERS, fieldName, "")
			}

		}
	}

	if len(field.InvalidChars) > 0 && hasStrVal {
		if strings.ContainsAny(strVal, field.InvalidChars) {
			httperror.NewFieldAPIError(httperror.INVALID_CHARACTERS, fieldName, "")
		}
	}

	return nil
}

func (b *Builder) convert(fieldType string, value interface{}, op Operation) (interface{}, error) {
	if value == nil {
		return value, nil
	}

	switch {
	case isMapType(fieldType):
		return b.convertMap(fieldType, value, op), nil
	case isArrayType(fieldType):
		return b.convertArray(fieldType, value, op), nil
	case isReferenceType(fieldType):
		return b.convertReferenceType(fieldType, value)
	}

	switch fieldType {
	case "json":
		return value, nil
	case "date":
		return convertString(value), nil
	case "boolean":
		return convertBool(value), nil
	case "enum":
		return convertString(value), nil
	case "int":
		return convertNumber(value)
	case "password":
		return convertString(value), nil
	case "string":
		return convertString(value), nil
	case "reference":
		return convertString(value), nil
	}

	return b.convertType(fieldType, value, op)
}

func (b *Builder) convertType(fieldType string, value interface{}, op Operation) (interface{}, error) {
	schema := b.Registry.GetSchema(fieldType)
	if schema == nil {
		return nil, httperror.NewAPIError(httperror.INVALID_TYPE, "Failed to find type "+fieldType)
	}

	mapValue, ok := value.(map[string]interface{})
	if !ok {
		return nil, httperror.NewAPIError(httperror.INVALID_FORMAT, fmt.Sprintf("Value can not be converted to type %s: %v", fieldType, value))
	}

	return b.Construct(schema, mapValue, op)
}

func convertNumber(value interface{}) (int64, error) {
	i, ok := value.(int64)
	if ok {
		return i, nil
	}
	return strconv.ParseInt(convertString(value), 10, 64)
}

func convertBool(value interface{}) bool {
	b, ok := value.(bool)
	if ok {
		return b
	}

	str := strings.ToLower(convertString(value))
	return str == "true" || str == "t" || str == "yes" || str == "y"
}

func convertString(value interface{}) string {
	return fmt.Sprint(value)
}

func (b *Builder) convertReferenceType(fieldType string, value interface{}) (string, error) {
	subType := fieldType[len("array[") : len(fieldType)-1]
	strVal := convertString(value)
	if !b.RefValidator.Validate(subType, strVal) {
		return "", httperror.NewAPIError(httperror.INVALID_REFERENCE, fmt.Sprintf("Not found type: %s id: %s", subType, strVal))
	}
	return strVal, nil
}

func (b *Builder) convertArray(fieldType string, value interface{}, op Operation) ([]interface{}, error) {
	sliceValue, ok := value.([]interface{})
	if !ok {
		return nil, nil
	}

	result := []interface{}{}
	subType := fieldType[len("array[") : len(fieldType)-1]

	for _, value := range sliceValue {
		val, err := b.convert(subType, value, op)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}

	return result, nil
}

func (b *Builder) convertMap(fieldType string, value interface{}, op Operation) (map[string]interface{}, error) {
	mapValue, ok := value.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	result := map[string]interface{}{}
	subType := fieldType[len("map[") : len(fieldType)-1]

	for key, value := range mapValue {
		val, err := b.convert(subType, value, op)
		if err != nil {
			return nil, httperror.WrapAPIError(err, httperror.INVALID_FORMAT, err.Error())
		}
		result[key] = val
	}

	return result, nil
}

func isMapType(fieldType string) bool {
	return strings.HasPrefix(fieldType, "map[") && strings.HasSuffix(fieldType, "]")
}

func isArrayType(fieldType string) bool {
	return strings.HasPrefix(fieldType, "array[") && strings.HasSuffix(fieldType, "]")
}

func isReferenceType(fieldType string) bool {
	return strings.HasPrefix(fieldType, "reference[") && strings.HasSuffix(fieldType, "]")
}

func fieldMatchesOp(field client.Field, op Operation) bool {
	switch op {
	case Create:
		return field.Create
	case Update:
		return field.Update
	default:
		return false
	}
}
