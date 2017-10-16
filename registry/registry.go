package registry

import "github.com/rancher/go-rancher/v3"

type SchemaRegistry interface {
	GetSchema(name string) *client.Schema
}
