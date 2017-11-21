package handler

import (
	"github.com/rancher/norman/parse"
	"github.com/rancher/norman/parse/builder"
	"github.com/rancher/norman/types"
)

func ParseAndValidateBody(apiContext *types.APIContext) (map[string]interface{}, error) {
	data, err := parse.Body(apiContext.Request)
	if err != nil {
		return nil, err
	}

	b := builder.NewBuilder(apiContext)

	data, err = b.Construct(apiContext.Schema, data, builder.Create)
	validator := apiContext.Schema.Validator
	if validator != nil {
		if err := validator(apiContext, data); err != nil {
			return nil, err
		}
	}

	return data, nil
}
