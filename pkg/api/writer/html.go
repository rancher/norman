package writer

import (
	"strings"

	"github.com/rancher/norman/pkg/types"
)

const (
	JSURL          = "https://releases.rancher.com/api-ui/%API_UI_VERSION%/ui.min.js"
	CSSURL         = "https://releases.rancher.com/api-ui/%API_UI_VERSION%/ui.min.css"
	DefaultVersion = "1.1.6"
)

var (
	start = `
<!DOCTYPE html>
<!-- If you are reading this, there is a good chance you would prefer sending an
"Accept: application/json" header and receiving actual JSON responses. -->
<link rel="stylesheet" type="text/css" href="%CSSURL%" />
<script src="%JSURL%"></script>
<script>
var user = "admin";
var curlUser='${CATTLE_ACCESS_KEY}:${CATTLE_SECRET_KEY}';
var schemas="%SCHEMAS%";
var data =
`
	end = []byte(`</script>
`)
)

type StringGetter func() string

type HTMLResponseWriter struct {
	EncodingResponseWriter
	CSSURL       StringGetter
	JSURL        StringGetter
	APIUIVersion StringGetter
}

func (h *HTMLResponseWriter) start(apiOp *types.APIRequest, code int, obj interface{}) {
	AddCommonResponseHeader(apiOp)
	apiOp.Response.Header().Set("content-type", "text/html")
	apiOp.Response.WriteHeader(code)
}

func (h *HTMLResponseWriter) Write(apiOp *types.APIRequest, code int, obj interface{}) {
	h.start(apiOp, code, obj)
	schemaSchema := apiOp.Schemas.Schema("schema")
	headerString := start
	if schemaSchema != nil {
		headerString = strings.Replace(headerString, "%SCHEMAS%", apiOp.URLBuilder.Collection(schemaSchema), 1)
	}
	var jsurl, cssurl string
	if h.CSSURL != nil && h.JSURL != nil && h.CSSURL() != "" && h.JSURL() != "" {
		jsurl = h.JSURL()
		cssurl = h.CSSURL()
	} else if h.APIUIVersion != nil && h.APIUIVersion() != "" {
		jsurl = strings.Replace(JSURL, "%API_UI_VERSION%", h.APIUIVersion(), 1)
		cssurl = strings.Replace(CSSURL, "%API_UI_VERSION%", h.APIUIVersion(), 1)
	} else {
		jsurl = strings.Replace(JSURL, "%API_UI_VERSION%", DefaultVersion, 1)
		cssurl = strings.Replace(CSSURL, "%API_UI_VERSION%", DefaultVersion, 1)
	}
	headerString = strings.Replace(headerString, "%JSURL%", jsurl, 1)
	headerString = strings.Replace(headerString, "%CSSURL%", cssurl, 1)

	apiOp.Response.Write([]byte(headerString))
	h.Body(apiOp, apiOp.Response, obj)
	if schemaSchema != nil {
		apiOp.Response.Write(end)
	}
}
