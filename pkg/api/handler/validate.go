package handler

import (
	"github.com/rancher/norman/pkg/parse"
	builder2 "github.com/rancher/norman/pkg/parse/builder"
	"github.com/rancher/norman/pkg/types"
)

func ParseAndValidateBody(apiOp *types.APIOperation, create bool) (map[string]interface{}, error) {
	data, err := parse.Body(apiOp.Request)
	if err != nil {
		return nil, err
	}

	b := builder2.NewBuilder(apiOp)

	op := builder2.Create
	if !create {
		op = builder2.Update
	}
	if apiOp.Schema.InputFormatter != nil {
		err = apiOp.Schema.InputFormatter(apiOp, apiOp.Schema, data, create)
		if err != nil {
			return nil, err
		}
	}
	data, err = b.Construct(apiOp.Schema, data, op)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ParseAndValidateActionBody(apiOp *types.APIOperation, actionInputSchema *types.Schema) (map[string]interface{}, error) {
	data, err := parse.Body(apiOp.Request)
	if err != nil {
		return nil, err
	}

	b := builder2.NewBuilder(apiOp)

	op := builder2.Create
	data, err = b.Construct(actionInputSchema, data, op)
	if err != nil {
		return nil, err
	}

	return data, nil
}
