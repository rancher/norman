package handlers

import (
	"net/http"

	"github.com/rancher/norman/types"
)

func CreateHandler(request *types.APIContext) error {
	var err error

	validator := request.Schema.Validator
	if validator != nil {
		if err := validator(request, request.Body); err != nil {
			return err
		}
	}

	data := request.Body

	store := request.Schema.Store
	if store != nil {
		data, err = store.Create(request, request.Schema, data)
		if err != nil {
			return err
		}
	}

	request.WriteResponse(http.StatusCreated, data)
	return nil
}
