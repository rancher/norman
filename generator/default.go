package generator

import (
	"path"
	"strings"

	"github.com/rancher/norman/types"
)

var (
	baseCattle = "client"
	baseK8s    = "apis"
)

func DefaultGenerate(schemas *types.Schemas, pkgPath string, publicAPI bool, backendTypes map[string]bool) error {
	version := getVersion(schemas)
	group := strings.Split(version.Group, ".")[0]

	cattleOutputPackage := path.Join(pkgPath, baseCattle, group, version.Version)
	if !publicAPI {
		cattleOutputPackage = ""
	}
	k8sOutputPackage := path.Join(pkgPath, baseK8s, version.Group, version.Version)

	if err := Generate(schemas, backendTypes, cattleOutputPackage, k8sOutputPackage); err != nil {
		return err
	}

	return nil
}

func getVersion(schemas *types.Schemas) *types.APIVersion {
	var version types.APIVersion
	for _, schema := range schemas.Schemas() {
		if version.Group == "" {
			version = schema.Version
			continue
		}
		if version.Group != schema.Version.Group ||
			version.Version != schema.Version.Version {
			panic("schema set contains two APIVersions")
		}
	}

	return &version
}
