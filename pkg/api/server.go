package api

import (
	"net/http"
	"sync"

	access2 "github.com/rancher/norman/pkg/api/access"
	builtin2 "github.com/rancher/norman/pkg/api/builtin"
	handler2 "github.com/rancher/norman/pkg/api/handler"
	writer2 "github.com/rancher/norman/pkg/api/writer"
	"github.com/rancher/norman/pkg/authorization"
	"github.com/rancher/norman/pkg/httperror"
	handler3 "github.com/rancher/norman/pkg/httperror/handler"
	"github.com/rancher/norman/pkg/parse"

	"github.com/rancher/norman/pkg/store/wrapper"
	"github.com/rancher/norman/pkg/types"
)

type Verb string

var (
	List   = Verb("list")
	Get    = Verb("get")
	Delete = Verb("delete")
	Create = Verb("create")
	Update = Verb("update")
)

type StoreWrapper func(types.Store) types.Store

type RequestHandler interface {
	http.Handler

	Handle(apiOp *types.APIOperation)
}

type Server struct {
	initBuiltin      sync.Once
	IgnoreBuiltin    bool
	ResponseWriters  map[string]ResponseWriter
	Schemas          *types.Schemas
	QueryFilter      types.QueryFilter
	StoreWrapper     StoreWrapper
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

func NewAPIServer() *Server {
	s := &Server{
		DefaultNamespace: "default",
		Schemas:          types.NewSchemas(),
		ResponseWriters: map[string]ResponseWriter{
			"json": &writer2.EncodingResponseWriter{
				ContentType: "application/json",
				Encoder:     types.JSONEncoder,
			},
			"html": &writer2.HTMLResponseWriter{
				EncodingResponseWriter: writer2.EncodingResponseWriter{
					Encoder:     types.JSONEncoder,
					ContentType: "application/json",
				},
			},
			"yaml": &writer2.EncodingResponseWriter{
				ContentType: "application/yaml",
				Encoder:     types.YAMLEncoder,
			},
		},
		AccessControl: &authorization.AllAccess{},
		Defaults: Defaults{
			CreateHandler: handler2.CreateHandler,
			DeleteHandler: handler2.DeleteHandler,
			UpdateHandler: handler2.UpdateHandler,
			ListHandler:   handler2.ListHandler,
			ErrorHandler:  handler3.ErrorHandler,
		},
		StoreWrapper: wrapper.Wrap,
		QueryFilter:  handler2.QueryFilter,
		Parser:       parse.Parse,
		URLParser:    parse.MuxURLParser,
	}

	return s
}

func (s *Server) setDefaults(ctx *types.APIOperation) {
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
}

func (s *Server) AddSchemas(schemas *types.Schemas) error {
	if schemas.Err() != nil {
		return schemas.Err()
	}

	s.initBuiltin.Do(func() {
		if s.IgnoreBuiltin {
			return
		}
		for _, schema := range builtin2.Schemas.Schemas() {
			s.addSchema(*schema)
		}
	})

	for _, schema := range schemas.Schemas() {
		s.addSchema(*schema)
	}

	return s.Schemas.Err()
}

func (s *Server) addSchema(schema types.Schema) {
	s.setupDefaults(&schema)
	s.Schemas.AddSchema(schema)
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

	if schema.Store != nil && s.StoreWrapper != nil {
		schema.Store = s.StoreWrapper(schema.Store)
	}
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.Handle(&types.APIOperation{
		Namespaces: []string{s.DefaultNamespace},
		Request:    req,
		Response:   rw,
	})
}

func (s *Server) Handle(apiOp *types.APIOperation) {
	s.handle(apiOp, apiOp.Response, apiOp.Request, s.Parser)
}

func (s *Server) handle(apiOp *types.APIOperation, rw http.ResponseWriter, req *http.Request, parser parse.Parser) {
	if err := parser(apiOp, rw, req, s.Schemas, parse.MuxURLParser); err != nil {
		s.handleError(apiOp, err)
		return
	}

	if code, data, err := s.handleOp(apiOp); err != nil {
		s.handleError(apiOp, err)
	} else {
		apiOp.WriteResponse(code, data)
	}
}

func determineVerb(apiOp *types.APIOperation) Verb {
	if apiOp.Link != "" {
		return List
	}

	switch apiOp.Method {
	case http.MethodGet:
		if apiOp.Name == "" {
			return List
		} else {
			return Get
		}
	case http.MethodPost:
		return Create
	case http.MethodPut:
		return Update
	case http.MethodDelete:
		return Delete
	}

	return ""
}

func (s *Server) handleOp(apiOp *types.APIOperation) (int, interface{}, error) {
	s.setDefaults(apiOp)

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
		if data == nil {
			return http.StatusNoContent, data, err
		} else {
			return http.StatusOK, data, err
		}
	}

	return http.StatusNotFound, nil, httperror.NewAPIError(httperror.NotFound, "")
}

func handle(apiOp *types.APIOperation, custom types.RequestHandler, handler types.RequestHandler) (interface{}, error) {
	if custom != nil {
		return custom(apiOp, handler)
	} else if handler != nil {
		return handler(apiOp, nil)
	}
	return nil, httperror.NewAPIError(httperror.NotFound, "")
}

func handleAction(action *types.Action, context *types.APIOperation) error {
	if context.Name != "" {
		if err := access2.ByID(context, context.Type, context.Name, nil); err != nil {
			return err
		}
	}
	return context.Schema.ActionHandler(context.Action, action, context)
}

func (s *Server) handleError(apiOp *types.APIOperation, err error) {
	if apiOp.Schema == nil {
		s.Defaults.ErrorHandler(apiOp, err)
	} else if apiOp.Schema.ErrorHandler != nil {
		apiOp.Schema.ErrorHandler(apiOp, err)
	}
}

func (s *Server) CustomAPIUIResponseWriter(cssURL, jsURL, version writer2.StringGetter) {
	wi, ok := s.ResponseWriters["html"]
	if !ok {
		return
	}
	w, ok := wi.(*writer2.HTMLResponseWriter)
	if !ok {
		return
	}
	w.CSSURL = cssURL
	w.JSURL = jsURL
	w.APIUIVersion = version
}
