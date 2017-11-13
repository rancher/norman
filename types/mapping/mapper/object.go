package mapper

import "github.com/rancher/norman/types"

type Object struct {
	types.TypeMapper
}

func NewObject(mappers []types.Mapper) *Object {
	return &Object{
		TypeMapper: types.TypeMapper{
			Mappers: append(mappers,
				&Drop{"status"},
				&Embed{Field: "metadata"},
				&Embed{Field: "spec"},
			),
		},
	}
}
