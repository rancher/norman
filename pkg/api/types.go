package api

import "github.com/rancher/norman/v2/pkg/types"

type ResponseWriter interface {
	Write(apiOp *types.APIRequest, code int, obj interface{})
}
