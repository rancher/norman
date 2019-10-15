package parse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rancher/norman/pkg/types"

	"github.com/rancher/norman/pkg/httperror"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const reqMaxSize = (2 * 1 << 20) + 1

var bodyMethods = map[string]bool{
	http.MethodPut:  true,
	http.MethodPost: true,
}

type Decode func(interface{}) error

func ReadBody(req *http.Request) (types.APIObject, error) {
	if !bodyMethods[req.Method] {
		return types.APIObject{}, nil
	}

	decode := getDecoder(req, io.LimitReader(req.Body, maxFormSize))

	data := map[string]interface{}{}
	if err := decode(&data); err != nil {
		return types.APIObject{}, httperror.NewAPIError(httperror.InvalidBodyContent,
			fmt.Sprintf("Failed to parse body: %v", err))
	}

	return types.ToAPI(data), nil
}

func getDecoder(req *http.Request, reader io.Reader) Decode {
	if req.Header.Get("Content-type") == "application/yaml" {
		return yaml.NewYAMLToJSONDecoder(reader).Decode
	}
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	return decoder.Decode
}
