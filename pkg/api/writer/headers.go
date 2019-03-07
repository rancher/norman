package writer

import (
	"github.com/rancher/norman/pkg/types"
)

func AddCommonResponseHeader(apiOp *types.APIOperation) error {
	addExpires(apiOp)
	return addSchemasHeader(apiOp)
}

func addSchemasHeader(apiOp *types.APIOperation) error {
	schema := apiOp.Schemas.Schema("schema")
	if schema == nil {
		return nil
	}

	apiOp.Response.Header().Set("X-Api-Schemas", apiOp.URLBuilder.Collection(schema))
	return nil
}

func addExpires(apiOp *types.APIOperation) {
	apiOp.Response.Header().Set("Expires", "Wed 24 Feb 1982 18:42:00 GMT")
}
