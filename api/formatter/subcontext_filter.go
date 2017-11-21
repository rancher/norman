package formatter

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
)

func SubContextFormatter(apiContext *types.APIContext, resource *types.RawResource) {
	if resource.Schema.SubContext == "" {
		return
	}

	ref := convert.ToReference(resource.Schema.ID)
	fullRef := convert.ToFullReference(resource.Schema.Version.Path, resource.Schema.ID)

outer:
	for _, schema := range apiContext.Schemas.Schemas() {
		for _, field := range schema.ResourceFields {
			if (field.Type == ref || field.Type == fullRef) && schema.Version.SubContexts[resource.Schema.SubContext] {
				resource.Links[schema.PluralName] = apiContext.URLBuilder.SubContextCollection(resource.Schema, resource.ID, schema)
				continue outer
			}
		}
	}
}
