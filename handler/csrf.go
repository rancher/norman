package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/rancher/norman/httperror"
)

const (
	csrfCookie = "CSRF"
	csrfHeader = "X-API-CSRF"
)

func CheckCSRF(rw http.ResponseWriter, req *http.Request) error {
	if !IsBrowser(req, false) {
		return nil
	}

	cookie, err := req.Cookie(csrfCookie)
	if err != nil {
		return httperror.NewAPIError(httperror.INVALID_CSRF_TOKEN, "Failed to parse cookies")
	}

	if cookie == nil {
		bytes := make([]byte, 5)
		_, err := rand.Read(bytes)
		if err != nil {
			return httperror.WrapAPIError(err, httperror.SERVER_ERROR, "Failed in CSRF processing")
		}

		cookie = &http.Cookie{
			Name:  csrfCookie,
			Value: hex.EncodeToString(bytes),
		}
	} else if req.Method != http.MethodGet {
		/*
		 * Very important to use request.getMethod() and not httpRequest.getMethod(). The client can override the HTTP method with _method
		 */
		if cookie.Value == req.Header.Get(csrfHeader) {
			// Good
		} else if cookie.Value == req.URL.Query().Get(csrfCookie) {
			// Good
		} else {
			return httperror.NewAPIError(httperror.INVALID_CSRF_TOKEN, "Invalid CSRF token")
		}
	}

	cookie.Path = "/"
	http.SetCookie(rw, cookie)
	return nil
}
