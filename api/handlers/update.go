package handlers

import (
	"net/http"

	"github.com/rancher/norman/types"
)

func UpdateHandler(request *types.APIContext) error {
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
		data, err = store.Update(request, request.Schema, data, request.ID)
		if err != nil {
			return err
		}
	}

	request.WriteResponse(http.StatusOK, data)
	return nil
}
