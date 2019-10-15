package handler

import (
	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/parse"
	"github.com/rancher/norman/pkg/types"
)

func ListHandler(request *types.APIRequest) (types.APIObject, error) {
	var (
		err  error
		data types.APIObject
	)

	if request.Name == "" {
		if err := request.AccessControl.CanList(request, request.Schema); err != nil {
			return types.APIObject{}, err
		}
	} else {
		if err := request.AccessControl.CanGet(request, request.Schema); err != nil {
			return types.APIObject{}, err
		}
	}

	store := request.Schema.Store
	if store == nil {
		return types.APIObject{}, httperror.NewAPIError(httperror.NotFound, "no store found")
	}

	if request.Name == "" {
		opts := parse.QueryOptions(request, request.Schema)
		// Save the pagination on the context so it's not reset later
		request.Pagination = opts.Pagination
		data, err = store.List(request, request.Schema, &opts)
		data = request.Filter(&opts, request.Schema, data)
		if data.IsNil() {
			data = types.ToAPI([]interface{}{})
		}
	} else if request.Link == "" {
		data, err = store.ByID(request, request.Schema, request.Name)
		data = request.Filter(nil, request.Schema, data)
	} else {
		_, err = store.ByID(request, request.Schema, request.Name)
		if err != nil {
			return types.APIObject{}, err
		}
		return request.Schema.LinkHandler(request)
	}

	if err != nil {
		return types.APIObject{}, err
	}

	return data, nil
}
