package api

import (
	"context"
	"net/http"

	"github.com/rancher/norman/api/builtin"
	"github.com/rancher/norman/api/handlers"
	"github.com/rancher/norman/api/writer"
	"github.com/rancher/norman/authorization"
	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/parse"
	"github.com/rancher/norman/parse/builder"
	"github.com/rancher/norman/types"
)

type Parser func(rw http.ResponseWriter, req *http.Request) (*types.APIContext, error)

type Server struct {
	IgnoreBuiltin   bool
	Parser          Parser
	ResponseWriters map[string]ResponseWriter
	schemas         *types.Schemas
	Defaults        Defaults
}

type Defaults struct {
	ActionHandler types.ActionHandler
	ListHandler   types.RequestHandler
	LinkHandler   types.RequestHandler
	CreateHandler types.RequestHandler
	DeleteHandler types.RequestHandler
	UpdateHandler types.RequestHandler
	Store         types.Store
	ErrorHandler  types.ErrorHandler
}

func NewAPIServer() *Server {
	s := &Server{
		schemas: types.NewSchemas(),
		ResponseWriters: map[string]ResponseWriter{
			"json": &writer.JSONResponseWriter{},
			"html": &writer.HTMLResponseWriter{},
		},
		Defaults: Defaults{
			CreateHandler: handlers.CreateHandler,
			DeleteHandler: handlers.DeleteHandler,
			UpdateHandler: handlers.UpdateHandler,
			ListHandler:   handlers.ListHandler,
			LinkHandler: func(*types.APIContext) error {
				return httperror.NewAPIError(httperror.NOT_FOUND, "Link not found")
			},
			ErrorHandler: httperror.ErrorHandler,
		},
	}

	s.Parser = func(rw http.ResponseWriter, req *http.Request) (*types.APIContext, error) {
		ctx, err := parse.Parse(rw, req, s.schemas)
		ctx.ResponseWriter = s.ResponseWriters[ctx.ResponseFormat]
		if ctx.ResponseWriter == nil {
			ctx.ResponseWriter = s.ResponseWriters["json"]
		}

		ctx.AccessControl = &authorization.AllAccess{}

		return ctx, err
	}

	return s
}

func (s *Server) Start(ctx context.Context) error {
	return s.addBuiltins(ctx)
}

func (s *Server) AddSchemas(schemas *types.Schemas) error {
	if schemas.Err() != nil {
		return schemas.Err()
	}

	for _, schema := range schemas.Schemas() {
		s.setupDefaults(schema)
		s.schemas.AddSchema(schema)
	}

	return s.schemas.Err()
}

func (s *Server) setupDefaults(schema *types.Schema) {
	if schema.ActionHandler == nil {
		schema.ActionHandler = s.Defaults.ActionHandler
	}

	if schema.Store == nil {
		schema.Store = s.Defaults.Store
	}

	if schema.ListHandler == nil {
		schema.ListHandler = s.Defaults.ListHandler
	}

	if schema.LinkHandler == nil {
		schema.LinkHandler = s.Defaults.LinkHandler
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

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if apiResponse, err := s.handle(rw, req); err != nil {
		s.handleError(apiResponse, err)
	}
}

func (s *Server) handle(rw http.ResponseWriter, req *http.Request) (*types.APIContext, error) {
	apiRequest, err := s.Parser(rw, req)
	if err != nil {
		return apiRequest, err
	}

	if err := CheckCSRF(rw, req); err != nil {
		return apiRequest, err
	}

	if err := addCommonResponseHeader(apiRequest); err != nil {
		return apiRequest, err
	}

	action, err := ValidateAction(apiRequest)
	if err != nil {
		return apiRequest, err
	}

	if apiRequest.Schema == nil {
		return apiRequest, nil
	}

	b := builder.NewBuilder(apiRequest)

	if action == nil && apiRequest.Type != "" {
		var handler types.RequestHandler
		switch apiRequest.Method {
		case http.MethodGet:
			handler = apiRequest.Schema.ListHandler
			apiRequest.Body = nil
		case http.MethodPost:
			handler = apiRequest.Schema.CreateHandler
			apiRequest.Body, err = b.Construct(apiRequest.Schema, apiRequest.Body, builder.Create)
		case http.MethodPut:
			handler = apiRequest.Schema.UpdateHandler
			apiRequest.Body, err = b.Construct(apiRequest.Schema, apiRequest.Body, builder.Update)
		case http.MethodDelete:
			handler = apiRequest.Schema.DeleteHandler
			apiRequest.Body = nil
		}

		if err != nil {
			return apiRequest, err
		}

		if handler == nil {
			return apiRequest, httperror.NewAPIError(httperror.NOT_FOUND, "")
		}

		return apiRequest, handler(apiRequest)
	} else if action != nil {
		return apiRequest, handleAction(action, apiRequest)
	}

	return apiRequest, nil
}

func handleAction(action *types.Action, request *types.APIContext) error {
	return request.Schema.ActionHandler(request.Action, action, request)
}

func (s *Server) handleError(apiRequest *types.APIContext, err error) {
	if apiRequest.Schema == nil {
		s.Defaults.ErrorHandler(apiRequest, err)
	} else if apiRequest.Schema.ErrorHandler != nil {
		apiRequest.Schema.ErrorHandler(apiRequest, err)
	}
}

func (s *Server) addBuiltins(ctx context.Context) error {
	if s.IgnoreBuiltin {
		return nil
	}

	if err := s.AddSchemas(builtin.Schemas); err != nil {
		return err
	}

	return nil
}
