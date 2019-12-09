package parse

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rancher/norman/v2/pkg/types"
)

func MuxURLParser(rw http.ResponseWriter, req *http.Request, schemas *types.Schemas) (ParsedURL, error) {
	vars := mux.Vars(req)
	url := ParsedURL{
		Type:   vars["type"],
		Name:   vars["name"],
		Link:   vars["link"],
		Prefix: vars["prefix"],
		Method: req.Method,
		Action: vars["action"],
		Query:  req.URL.Query(),
	}

	return url, nil
}
