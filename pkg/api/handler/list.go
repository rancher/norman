package handler

import (
	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/parse"
	"github.com/rancher/norman/pkg/types"
)

func ListHandler(request *types.APIOperation, next types.RequestHandler) (interface{}, error) {
	var (
		err  error
		data interface{}
	)

	if request.Name == "" {
		if err := request.AccessControl.CanList(request, request.Schema); err != nil {
			return nil, err
		}
	} else {
		if err := request.AccessControl.CanGet(request, request.Schema); err != nil {
			return nil, err
		}
	}

	store := request.Schema.Store
	if store == nil {
		return nil, httperror.NewAPIError(httperror.NotFound, "no store found")
	}

	if request.Name == "" {
		opts := parse.QueryOptions(request, request.Schema)
		// Save the pagination on the context so it's not reset later
		request.Pagination = opts.Pagination
		data, err = store.List(request, request.Schema, &opts)
	} else if request.Link == "" {
		data, err = store.ByID(request, request.Schema, request.Name)
	} else {
		_, err = store.ByID(request, request.Schema, request.Name)
		if err != nil {
			return nil, err
		}
		return request.Schema.LinkHandler(request, nil)
	}

	if err != nil {
		return nil, err
	}

	return data, nil
}
