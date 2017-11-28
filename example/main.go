package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rancher/norman/api"
	"github.com/rancher/norman/store/crd"
	"github.com/rancher/norman/types"
	"k8s.io/client-go/tools/clientcmd"
)

type Foo struct {
	types.Resource
	Name     string `json:"name"`
	Foo      string `json:"foo"`
	SubThing Baz    `json:"subThing"`
}

type Baz struct {
	Name string `json:"name"`
}

var (
	version = types.APIVersion{
		Version: "v1",
		Group:   "io.cattle.core.example",
		Path:    "/example/v1",
	}

	Schemas = types.NewSchemas()
)

func main() {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err)
	}

	store, err := crd.NewCRDStoreFromConfig(*kubeConfig)
	if err != nil {
		panic(err)
	}

	Schemas.MustImportAndCustomize(&version, Foo{}, func(schema *types.Schema) {
		schema.Store = store
	})

	server := api.NewAPIServer()
	if err := server.AddSchemas(Schemas); err != nil {
		panic(err)
	}

	fmt.Println("Listening on 0.0.0.0:1234")
	http.ListenAndServe("0.0.0.0:1234", server)
}
