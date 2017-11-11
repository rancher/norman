package handlers

import (
	"net/http"

	"github.com/rancher/norman/parse"
	"github.com/rancher/norman/types"
)

func ListHandler(request *types.APIContext) error {
	var (
		err  error
		data interface{}
	)

	store := request.Schema.Store
	if store == nil {
		return nil
	}

	if request.ID == "" {
		request.QueryOptions = parse.QueryOptions(request.Request, request.Schema)
		data, err = store.List(request, request.Schema, request.QueryOptions)
	} else if request.Link == "" {
		data, err = store.ByID(request, request.Schema, request.ID)
	} else {
		return request.Schema.LinkHandler(request)
	}

	if err != nil {
		return err
	}

	request.WriteResponse(http.StatusOK, data)
	return nil
}
