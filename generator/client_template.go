package generator

var clientTemplate = `package client

import (
	"github.com/rancher/norman/clientbase"
)

type Client struct {
    clientbase.APIBaseClient

	{{range .schemas}}
    {{- if . | hasGet }}{{.ID | capitalize}} {{.ID | capitalize}}Operations
{{end}}{{end}}}

func NewClient(opts *clientbase.ClientOpts) (*Client, error) {
	baseClient, err := clientbase.NewAPIClient(opts)
	if err != nil {
		return nil, err
	}

	client := &Client{
        APIBaseClient: baseClient,
    }

    {{range .schemas}}
    {{- if . | hasGet }}client.{{.ID | capitalize}} = new{{.ID | capitalize}}Client(client)
{{end}}{{end}}
	return client, nil
}
`
