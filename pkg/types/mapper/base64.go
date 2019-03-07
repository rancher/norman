package mapper

import (
	"encoding/base64"
	"strings"

	"github.com/rancher/norman/pkg/types"
	convert2 "github.com/rancher/norman/pkg/types/convert"
	values2 "github.com/rancher/norman/pkg/types/values"
)

type Base64 struct {
	Field            string
	IgnoreDefinition bool
	Separator        string
}

func (m Base64) FromInternal(data map[string]interface{}) {
	if v, ok := values2.RemoveValue(data, strings.Split(m.Field, m.getSep())...); ok {
		str := convert2.ToString(v)
		if str == "" {
			return
		}

		newData, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			log.Errorf("failed to base64 decode data")
		}

		values2.PutValue(data, string(newData), strings.Split(m.Field, m.getSep())...)
	}
}

func (m Base64) ToInternal(data map[string]interface{}) error {
	if v, ok := values2.RemoveValue(data, strings.Split(m.Field, m.getSep())...); ok {
		str := convert2.ToString(v)
		if str == "" {
			return nil
		}

		newData := base64.StdEncoding.EncodeToString([]byte(str))
		values2.PutValue(data, newData, strings.Split(m.Field, m.getSep())...)
	}

	return nil
}

func (m Base64) ModifySchema(s *types.Schema, schemas *types.Schemas) error {
	if !m.IgnoreDefinition {
		if err := ValidateField(m.Field, s); err != nil {
			return err
		}
	}

	return nil
}

func (m Base64) getSep() string {
	if m.Separator == "" {
		return "/"
	}
	return m.Separator
}
