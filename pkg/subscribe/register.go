package subscribe

import (
	"net/http"

	"github.com/rancher/norman/pkg/types"
)

func Register(schemas *types.Schemas) {
	schemas.MustImportAndCustomize(Subscribe{}, func(schema *types.Schema) {
		schema.CollectionMethods = []string{http.MethodGet}
		schema.ResourceMethods = []string{}
		schema.ListHandler = Handler
		schema.PluralName = "subscribe"
	})
}
