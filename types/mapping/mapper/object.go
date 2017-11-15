package mapper

import "github.com/rancher/norman/types"

type Object struct {
	types.TypeMapper
}

func NewObject(mappers ...types.Mapper) *Object {
	return &Object{
		TypeMapper: types.TypeMapper{
			Mappers: append([]types.Mapper{
				&Embed{Field: "metadata"},
				&Embed{Field: "spec"},
				&ReadOnly{"status"},
			}, mappers...),
		},
	}
}
