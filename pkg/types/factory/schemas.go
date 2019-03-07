package factory

import (
	"github.com/rancher/norman/pkg/types"
	mapper2 "github.com/rancher/norman/pkg/types/mapper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Schemas() *types.Schemas {
	s := types.NewSchemas()
	s.DefaultMappers = func() []types.Mapper {
		return []types.Mapper{
			mapper2.NewObject(),
		}
	}
	s.DefaultPostMappers = func() []types.Mapper {
		return []types.Mapper{
			&mapper2.RenameReference{},
		}
	}
	s.AddMapperForType(v1.ObjectMeta{}, mapper2.NewMetadataMapper())
	return s
}
