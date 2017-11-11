package parse

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/rancher/norman/api/builtin"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/urlbuilder"
)

const (
	maxFormSize = 2 * 1 << 20
)

var (
	multiSlashRegexp = regexp.MustCompile("//+")
	allowedFormats   = map[string]bool{
		"html": true,
		"json": true,
	}
)

func Parse(rw http.ResponseWriter, req *http.Request, schemas *types.Schemas) (*types.APIContext, error) {
	var err error

	result := &types.APIContext{
		Request:  req,
		Response: rw,
	}

	// The response format is guarenteed to be set even in the event of an error
	result.ResponseFormat = parseResponseFormat(req)
	result.Version = parseVersion(schemas, req.URL.Path)
	result.Schemas = schemas

	if result.Version == nil {
		result.Method = http.MethodGet
		result.URLBuilder, err = urlbuilder.New(req, types.APIVersion{}, result.Schemas)
		result.Type = "apiRoot"
		result.Schema = &builtin.APIRoot
		return result, nil
	}

	result.Method = parseMethod(req)
	result.Action, result.Method = parseAction(req, result.Method)
	result.Body, err = parseBody(req)
	if err != nil {
		return result, err
	}

	result.URLBuilder, err = urlbuilder.New(req, *result.Version, result.Schemas)
	if err != nil {
		return result, err
	}

	parsePath(result, req)

	if result.Schema == nil {
		result.Method = http.MethodGet
		result.Type = "apiRoot"
		result.Schema = &builtin.APIRoot
		result.ID = result.Version.Path
		return result, nil
	}

	if err := ValidateMethod(result); err != nil {
		return result, err
	}

	return result, nil
}

func parseSubContext(parts []string, apiRequest *types.APIContext) []string {
	subContext := ""
	apiRequest.SubContext = map[string]interface{}{}

	for len(parts) > 3 && apiRequest.Version != nil {
		resourceType := parts[1]
		resourceID := parts[2]

		if !apiRequest.Version.SubContexts[resourceType] {
			break
		}

		if apiRequest.ReferenceValidator != nil && !apiRequest.ReferenceValidator.Validate(resourceType, resourceID) {
			return parts
		}

		apiRequest.SubContext[resourceType] = resourceID
		subContext = subContext + "/" + resourceType + "/" + resourceID
		parts = append(parts[:1], parts[3:]...)
	}

	if subContext != "" {
		apiRequest.URLBuilder.SetSubContext(subContext)
	}

	return parts
}

func parsePath(apiRequest *types.APIContext, request *http.Request) {
	if apiRequest.Version == nil {
		return
	}

	path := request.URL.Path
	path = multiSlashRegexp.ReplaceAllString(path, "/")

	versionPrefix := apiRequest.Version.Path
	if !strings.HasPrefix(path, versionPrefix) {
		return
	}

	parts := strings.Split(path[len(versionPrefix):], "/")
	parts = parseSubContext(parts, apiRequest)

	if len(parts) > 4 {
		return
	}

	typeName := safeIndex(parts, 1)
	id := safeIndex(parts, 2)
	link := safeIndex(parts, 3)

	if typeName == "" {
		return
	}

	schema := apiRequest.Schemas.Schema(apiRequest.Version, typeName)
	if schema == nil {
		return
	}

	apiRequest.Schema = schema
	apiRequest.Type = schema.ID

	if id == "" {
		return
	}

	apiRequest.ID = id
	apiRequest.Link = link
}

func safeIndex(slice []string, index int) string {
	if index >= len(slice) {
		return ""
	}
	return slice[index]
}

func parseResponseFormat(req *http.Request) string {
	format := req.URL.Query().Get("_format")

	if format != "" {
		format = strings.TrimSpace(strings.ToLower(format))
	}

	/* Format specified */
	if allowedFormats[format] {
		return format
	}

	// User agent has Mozilla and browser accepts */*
	if IsBrowser(req, true) {
		return "html"
	} else {
		return "json"
	}
}

func parseMethod(req *http.Request) string {
	method := req.URL.Query().Get("_method")
	if method == "" {
		method = req.Method
	}
	return method
}

func parseAction(req *http.Request, method string) (string, string) {
	if req.Method != http.MethodPost {
		return "", method
	}

	action := req.URL.Query().Get("action")
	if action == "remove" {
		return "", http.MethodDelete
	}

	return action, method
}

func parseVersion(schemas *types.Schemas, path string) *types.APIVersion {
	path = multiSlashRegexp.ReplaceAllString(path, "/")
	for _, version := range schemas.Versions() {
		if version.Path == "" {
			continue
		}
		if strings.HasPrefix(path, version.Path) {
			return &version
		}
	}

	return nil
}

func parseBody(req *http.Request) (map[string]interface{}, error) {
	req.ParseMultipartForm(maxFormSize)
	if req.MultipartForm != nil {
		return valuesToBody(req.MultipartForm.Value), nil
	}

	if req.Form != nil && len(req.Form) > 0 {
		return valuesToBody(map[string][]string(req.Form)), nil
	}

	return ReadBody(req)
}

func valuesToBody(input map[string][]string) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range input {
		result[k] = v
	}
	return result
}
