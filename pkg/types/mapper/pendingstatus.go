package mapper

import (
	"github.com/rancher/norman/pkg/types"
	convert2 "github.com/rancher/norman/pkg/types/convert"
	values2 "github.com/rancher/norman/pkg/types/values"
)

type PendingStatus struct {
}

func (s PendingStatus) FromInternal(data map[string]interface{}) {
	if data == nil {
		return
	}

	if data["state"] != "active" {
		return
	}

	conditions := convert2.ToMapSlice(values2.GetValueN(data, "status", "conditions"))
	if len(conditions) > 0 {
		return
	}

	data["state"] = "pending"
}

func (s PendingStatus) ToInternal(data map[string]interface{}) error {
	return nil
}

func (s PendingStatus) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return nil
}
