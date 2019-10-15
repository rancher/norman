package api

import (
	"net/http"

	"github.com/rancher/norman/pkg/api/access"
	"github.com/rancher/norman/pkg/api/handler"
	"github.com/rancher/norman/pkg/api/writer"
	"github.com/rancher/norman/pkg/authorization"
	"github.com/rancher/norman/pkg/httperror"
	errhandler "github.com/rancher/norman/pkg/httperror/handler"
	"github.com/rancher/norman/pkg/parse"
	"github.com/rancher/norman/pkg/types"
	"github.com/rancher/wrangler/pkg/merr"
)

type Verb string

var (
	List   = Verb("list")
	Get    = Verb("get")
	Delete = Verb("delete")
	Create = Verb("create")
	Update = Verb("update")
)

type RequestHandler interface {
	http.Handler

	GetSchemas() *types.Schemas
	Handle(apiOp *types.APIRequest)
}

type Server struct {
	ResponseWriters  map[string]ResponseWriter
	Schemas          *types.Schemas
	QueryFilter      types.QueryFilter
	Defaults         Defaults
	DefaultNamespace string
	AccessControl    types.AccessControl
	Parser           parse.Parser
	URLParser        parse.URLParser
}

type Defaults struct {
	ListHandler   types.RequestHandler
	CreateHandler types.RequestHandler
	DeleteHandler types.RequestHandler
	UpdateHandler types.RequestHandler
	Store         types.Store
	ErrorHandler  types.ErrorHandler
}

func DefaultAPIServer() *Server {
	s := &Server{
		DefaultNamespace: "default",
		Schemas:          types.EmptySchemas(),
		ResponseWriters: map[string]ResponseWriter{
			"json": &writer.EncodingResponseWriter{
				ContentType: "application/json",
				Encoder:     types.JSONEncoder,
			},
			"html": &writer.HTMLResponseWriter{
				EncodingResponseWriter: writer.EncodingResponseWriter{
					Encoder:     types.JSONEncoder,
					ContentType: "application/json",
				},
			},
			"yaml": &writer.EncodingResponseWriter{
				ContentType: "application/yaml",
				Encoder:     types.YAMLEncoder,
			},
		},
		AccessControl: &authorization.AllAccess{},
		Defaults: Defaults{
			CreateHandler: handler.CreateHandler,
			DeleteHandler: handler.DeleteHandler,
			UpdateHandler: handler.UpdateHandler,
			ListHandler:   handler.ListHandler,
			ErrorHandler:  errhandler.ErrorHandler,
		},
		QueryFilter: handler.QueryFilter,
		Parser:      parse.Parse,
		URLParser:   parse.MuxURLParser,
	}

	return s
}

func (s *Server) setDefaults(ctx *types.APIRequest) {
	if ctx.ResponseWriter == nil {
		ctx.ResponseWriter = s.ResponseWriters[ctx.ResponseFormat]
		if ctx.ResponseWriter == nil {
			ctx.ResponseWriter = s.ResponseWriters["json"]
		}
	}

	if ctx.QueryFilter == nil {
		ctx.QueryFilter = s.QueryFilter
	}

	ctx.AccessControl = s.AccessControl

	if ctx.Schemas == nil {
		ctx.Schemas = s.Schemas
	}
}

func (s *Server) AddSchemas(schemas *types.Schemas) error {
	var errs []error

	for _, schema := range schemas.Schemas() {
		if err := s.addSchema(*schema); err != nil {
			errs = append(errs, err)
		}
	}

	return merr.NewErrors(errs...)
}

func (s *Server) addSchema(schema types.Schema) error {
	s.setupDefaults(&schema)
	return s.Schemas.AddSchema(schema)
}

func (s *Server) setupDefaults(schema *types.Schema) {
	if schema.Store == nil {
		schema.Store = s.Defaults.Store
	}

	if schema.ListHandler == nil {
		schema.ListHandler = s.Defaults.ListHandler
	}

	if schema.CreateHandler == nil {
		schema.CreateHandler = s.Defaults.CreateHandler
	}

	if schema.UpdateHandler == nil {
		schema.UpdateHandler = s.Defaults.UpdateHandler
	}

	if schema.DeleteHandler == nil {
		schema.DeleteHandler = s.Defaults.DeleteHandler
	}

	if schema.ErrorHandler == nil {
		schema.ErrorHandler = s.Defaults.ErrorHandler
	}
}

func (s *Server) GetSchemas() *types.Schemas {
	return s.Schemas
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.Handle(&types.APIRequest{
		Namespaces: []string{s.DefaultNamespace},
		Request:    req,
		Response:   rw,
	})
}

func (s *Server) Handle(apiOp *types.APIRequest) {
	s.handle(apiOp, apiOp.Response, apiOp.Request, s.Parser)
}

func (s *Server) handle(apiOp *types.APIRequest, rw http.ResponseWriter, req *http.Request, parser parse.Parser) {
	if err := parser(apiOp, parse.MuxURLParser); err != nil {
		// ensure defaults set so writer is assigned
		s.setDefaults(apiOp)
		s.handleError(apiOp, err)
		return
	}

	s.setDefaults(apiOp)

	if code, data, err := s.handleOp(apiOp); err != nil {
		s.handleError(apiOp, err)
	} else {
		apiOp.WriteResponse(code, data)
	}
}

func determineVerb(apiOp *types.APIRequest) Verb {
	if apiOp.Link != "" {
		return List
	}

	switch apiOp.Method {
	case http.MethodGet:
		if apiOp.Name == "" {
			return List
		}
		return Get
	case http.MethodPost:
		return Create
	case http.MethodPatch:
		return Update
	case http.MethodPut:
		return Update
	case http.MethodDelete:
		return Delete
	}

	return ""
}

func (s *Server) handleOp(apiOp *types.APIRequest) (int, interface{}, error) {
	if err := CheckCSRF(apiOp); err != nil {
		return 0, nil, err
	}

	action, err := ValidateAction(apiOp)
	if err != nil {
		return 0, nil, err
	}

	if apiOp.Schema == nil {
		return http.StatusNotFound, nil, nil
	}

	if action != nil {
		return http.StatusOK, nil, handleAction(action, apiOp)
	}

	switch determineVerb(apiOp) {
	case Get:
		fallthrough
	case List:
		data, err := handle(apiOp, apiOp.Schema.ListHandler, s.Defaults.ListHandler)
		return http.StatusOK, data, err
	case Update:
		data, err := handle(apiOp, apiOp.Schema.UpdateHandler, s.Defaults.UpdateHandler)
		return http.StatusOK, data, err
	case Create:
		data, err := handle(apiOp, apiOp.Schema.CreateHandler, s.Defaults.CreateHandler)
		return http.StatusCreated, data, err
	case Delete:
		data, err := handle(apiOp, apiOp.Schema.DeleteHandler, s.Defaults.DeleteHandler)
		if err == nil && data.IsNil() {
			return http.StatusNoContent, data, err
		}
		return http.StatusOK, data, err
	}

	return http.StatusNotFound, nil, httperror.NewAPIError(httperror.NotFound, "")
}

func handle(apiOp *types.APIRequest, custom types.RequestHandler, handler types.RequestHandler) (types.APIObject, error) {
	var (
		obj types.APIObject
		err error
	)
	if custom != nil {
		obj, err = custom(apiOp)
	} else if handler != nil {
		obj, err = handler(apiOp)
	}

	if err == nil && obj.IsNil() {
		return types.APIObject{}, httperror.NewAPIError(httperror.NotFound, "")
	}

	return obj, err
}

func handleAction(action *types.Action, context *types.APIRequest) error {
	if context.Name != "" {
		if err := access.ByID(context, context.Type, context.Name, nil); err != nil {
			return err
		}
	}
	return context.Schema.ActionHandler(context.Action, action, context)
}

func (s *Server) handleError(apiOp *types.APIRequest, err error) {
	if apiOp.Schema != nil && apiOp.Schema.ErrorHandler != nil {
		apiOp.Schema.ErrorHandler(apiOp, err)
	} else if s.Defaults.ErrorHandler != nil {
		s.Defaults.ErrorHandler(apiOp, err)
	}
}

func (s *Server) CustomAPIUIResponseWriter(cssURL, jsURL, version writer.StringGetter) {
	wi, ok := s.ResponseWriters["html"]
	if !ok {
		return
	}
	w, ok := wi.(*writer.HTMLResponseWriter)
	if !ok {
		return
	}
	w.CSSURL = cssURL
	w.JSURL = jsURL
	w.APIUIVersion = version
}
