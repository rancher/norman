package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/rancher/norman/server"
	"github.com/rancher/norman/types"
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
	if _, err := Schemas.Import(&version, Foo{}); err != nil {
		panic(err)
	}

	server, err := server.NewAPIServer(context.Background(), os.Getenv("KUBECONFIG"), Schemas)
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening on 0.0.0.0:1234")
	http.ListenAndServe("0.0.0.0:1234", server)
}
