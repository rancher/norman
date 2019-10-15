package access

import (
	"fmt"

	"github.com/rancher/norman/pkg/parse/builder"
	"github.com/rancher/norman/pkg/types"
	"github.com/rancher/norman/pkg/types/convert"
)

func ByID(context *types.APIRequest, typeName string, id string, into interface{}) error {
	schema := context.Schemas.Schema(typeName)
	if schema == nil {
		return fmt.Errorf("failed to find schema " + typeName)
	}

	item, err := schema.Store.ByID(context, schema, id)
	if err != nil {
		return err
	}

	b := builder.NewBuilder(context)

	item, err = b.Construct(schema, item, builder.List)
	if err != nil {
		return err
	}

	if into == nil {
		return nil
	}

	return convert.ToObj(item.Raw(), into)
}
