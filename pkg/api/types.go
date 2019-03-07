package api

import "github.com/rancher/norman/pkg/types"

type ResponseWriter interface {
	Write(apiOp *types.APIOperation, code int, obj interface{})
}
