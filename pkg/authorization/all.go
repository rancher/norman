package authorization

import (
	"net/http"

	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/types"
	"github.com/rancher/norman/pkg/types/slice"
)

type AllAccess struct {
}

func (*AllAccess) CanCreate(apiOp *types.APIRequest, schema *types.Schema) error {
	if slice.ContainsString(schema.CollectionMethods, http.MethodPost) {
		return nil
	}
	return httperror.NewAPIError(httperror.PermissionDenied, "can not create "+schema.ID)
}

func (*AllAccess) CanGet(apiOp *types.APIRequest, schema *types.Schema) error {
	if slice.ContainsString(schema.ResourceMethods, http.MethodGet) {
		return nil
	}
	return httperror.NewAPIError(httperror.PermissionDenied, "can not get "+schema.ID)
}

func (*AllAccess) CanList(apiOp *types.APIRequest, schema *types.Schema) error {
	if slice.ContainsString(schema.CollectionMethods, http.MethodGet) {
		return nil
	}
	return httperror.NewAPIError(httperror.PermissionDenied, "can not list "+schema.ID)
}

func (*AllAccess) CanUpdate(apiOp *types.APIRequest, obj types.APIObject, schema *types.Schema) error {
	if slice.ContainsString(schema.ResourceMethods, http.MethodPut) {
		return nil
	}
	return httperror.NewAPIError(httperror.PermissionDenied, "can not update "+schema.ID)
}

func (*AllAccess) CanDelete(apiOp *types.APIRequest, obj types.APIObject, schema *types.Schema) error {
	if slice.ContainsString(schema.ResourceMethods, http.MethodDelete) {
		return nil
	}
	return httperror.NewAPIError(httperror.PermissionDenied, "can not delete "+schema.ID)
}

func (*AllAccess) CanWatch(apiOp *types.APIRequest, schema *types.Schema) error {
	return nil
}

func (*AllAccess) Filter(apiOp *types.APIRequest, schema *types.Schema, obj types.APIObject) types.APIObject {
	return obj
}

func (*AllAccess) FilterList(apiOp *types.APIRequest, schema *types.Schema, obj types.APIObject) types.APIObject {
	return obj
}
