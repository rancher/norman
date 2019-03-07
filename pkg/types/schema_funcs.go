package types

import (
	"net/http"

	slice2 "github.com/rancher/norman/pkg/types/slice"

	"github.com/rancher/norman/pkg/httperror"
)

func (s *Schema) MustCustomizeField(name string, f func(f Field) Field) *Schema {
	field, ok := s.ResourceFields[name]
	if !ok {
		panic("Failed to find field " + name + " on schema " + s.ID)
	}
	s.ResourceFields[name] = f(field)
	return s
}

func (s *Schema) CanList(context *APIOperation) error {
	if context == nil {
		if slice2.ContainsString(s.CollectionMethods, http.MethodGet) {
			return nil
		}
		return httperror.NewAPIError(httperror.PermissionDenied, "can not list "+s.ID)
	}
	return context.AccessControl.CanList(context, s)
}

func (s *Schema) CanGet(context *APIOperation) error {
	if context == nil {
		if slice2.ContainsString(s.ResourceMethods, http.MethodGet) {
			return nil
		}
		return httperror.NewAPIError(httperror.PermissionDenied, "can not get "+s.ID)
	}
	return context.AccessControl.CanGet(context, s)
}

func (s *Schema) CanCreate(context *APIOperation) error {
	if context == nil {
		if slice2.ContainsString(s.CollectionMethods, http.MethodPost) {
			return nil
		}
		return httperror.NewAPIError(httperror.PermissionDenied, "can not create "+s.ID)
	}
	return context.AccessControl.CanCreate(context, s)
}

func (s *Schema) CanUpdate(context *APIOperation) error {
	if context == nil {
		if slice2.ContainsString(s.ResourceMethods, http.MethodPut) {
			return nil
		}
		return httperror.NewAPIError(httperror.PermissionDenied, "can not update "+s.ID)
	}
	return context.AccessControl.CanUpdate(context, nil, s)
}

func (s *Schema) CanDelete(context *APIOperation) error {
	if context == nil {
		if slice2.ContainsString(s.ResourceMethods, http.MethodDelete) {
			return nil
		}
		return httperror.NewAPIError(httperror.PermissionDenied, "can not delete "+s.ID)
	}
	return context.AccessControl.CanDelete(context, nil, s)
}
