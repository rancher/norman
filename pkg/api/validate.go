package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/parse"

	"github.com/rancher/norman/pkg/types"
)

const (
	csrfCookie = "CSRF"
	csrfHeader = "X-API-CSRF"
)

func ValidateAction(request *types.APIRequest) (*types.Action, error) {
	if request.Action == "" || request.Link != "" || request.Method != http.MethodPost {
		return nil, nil
	}

	actions := request.Schema.CollectionActions
	if request.Name != "" {
		actions = request.Schema.ResourceActions
	}

	action, ok := actions[request.Action]
	if !ok {
		return nil, httperror.NewAPIError(httperror.InvalidAction, fmt.Sprintf("Invalid action: %s", request.Action))
	}

	if request.Name != "" && request.ReferenceValidator != nil {
		resource := request.ReferenceValidator.Lookup(request.Type, request.Name)
		if resource == nil {
			return nil, httperror.NewAPIError(httperror.NotFound, fmt.Sprintf("Failed to find type: %s name: %s", request.Type, request.Name))
		}

		if _, ok := resource.Actions[request.Action]; !ok {
			return nil, httperror.NewAPIError(httperror.InvalidAction, fmt.Sprintf("Invalid action: %s", request.Action))
		}
	}

	return &action, nil
}

func CheckCSRF(apiOp *types.APIRequest) error {
	if !parse.IsBrowser(apiOp.Request, false) {
		return nil
	}

	cookie, err := apiOp.Request.Cookie(csrfCookie)
	if err == http.ErrNoCookie {
		bytes := make([]byte, 5)
		_, err := rand.Read(bytes)
		if err != nil {
			return httperror.WrapAPIError(err, httperror.ServerError, "Failed in CSRF processing")
		}

		cookie = &http.Cookie{
			Name:  csrfCookie,
			Value: hex.EncodeToString(bytes),
		}
	} else if err != nil {
		return httperror.NewAPIError(httperror.InvalidCSRFToken, "Failed to parse cookies")
	} else if apiOp.Method != http.MethodGet {
		/*
		 * Very important to use apiOp.Method and not apiOp.Request.Method. The client can override the HTTP method with _method
		 */
		if cookie.Value == apiOp.Request.Header.Get(csrfHeader) {
			// Good
		} else if cookie.Value == apiOp.Request.URL.Query().Get(csrfCookie) {
			// Good
		} else {
			return httperror.NewAPIError(httperror.InvalidCSRFToken, "Invalid CSRF token")
		}
	}

	cookie.Path = "/"
	http.SetCookie(apiOp.Response, cookie)
	return nil
}
