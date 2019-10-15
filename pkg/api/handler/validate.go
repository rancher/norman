package handler

import (
	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/parse"
	"github.com/rancher/norman/pkg/parse/builder"
	"github.com/rancher/norman/pkg/types"
)

func ParseAndValidateBody(apiOp *types.APIRequest, create bool) (types.APIObject, error) {
	data, err := parse.Body(apiOp.Request)
	if err != nil {
		return types.APIObject{}, err
	}

	b := builder.NewBuilder(apiOp)

	op := builder.Create
	if !create {
		op = builder.Update
	}
	if apiOp.Schema.InputFormatter != nil {
		err = apiOp.Schema.InputFormatter(apiOp, apiOp.Schema, data, create)
		if err != nil {
			return types.APIObject{}, err
		}
	}
	data, err = b.Construct(apiOp.Schema, data, op)
	if err != nil {
		return types.APIObject{}, err
	}

	return data, nil
}

func validateGet(apiOp *types.APIRequest, schema *types.Schema, id string) error {
	store := schema.Store
	if store == nil {
		return nil
	}

	existing, err := store.ByID(apiOp, schema, id)
	if err != nil {
		return err
	}

	if apiOp.Filter(nil, schema, existing).IsNil() {
		return httperror.NewAPIError(httperror.NotFound, "failed to find "+id)
	}

	return nil
}
