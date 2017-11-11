Norman
========

An API framework for Building [Rancher Style APIs](https://github.com/rancher/api-spec/) backed by K8s CustomResources.

## Building

`make`

## Example

Refer to `examples/`

```go
package main

import (
	"context"
	"net/http"

	"fmt"

	"os"

	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/norman/api/crd"
)

var (
	version = client.APIVersion{
		Version: "v1",
		Group:   "io.cattle.core.example",
		Path:    "/example/v1",
	}

	Foo = client.Schema{
		ID:      "foo",
		Version: version,
		ResourceFields: map[string]*client.Field{
			"foo": {
				Type:   "string",
				Create: true,
				Update: true,
			},
			"name": {
				Type:     "string",
				Create:   true,
				Required: true,
			},
		},
	}

	Schemas = client.NewSchemas().
		AddSchema(&Foo)
)

func main() {
	server, err := crd.NewAPIServer(context.Background(), os.Getenv("KUBECONFIG"), Schemas)
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening on 0.0.0.0:1234")
	http.ListenAndServe("0.0.0.0:1234", server)
}
```


## License
Copyright (c) 2014-2017 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
