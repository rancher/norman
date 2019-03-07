package types

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"sync"

	convert2 "github.com/rancher/norman/pkg/types/convert"

	"github.com/rancher/wrangler/pkg/name"
)

type SchemaCollection struct {
	Data []Schema
}

type SchemasInitFunc func(*Schemas) *Schemas

type MappersFactory func() []Mapper

type Schemas struct {
	sync.Mutex
	processingTypes    map[reflect.Type]*Schema
	typeNames          map[reflect.Type]string
	schemasByID        map[string]*Schema
	mappers            map[string][]Mapper
	embedded           map[string]*Schema
	DefaultMappers     MappersFactory
	DefaultPostMappers MappersFactory
	schemas            []*Schema
	errors             []error
}

func NewSchemas() *Schemas {
	return &Schemas{
		processingTypes: map[reflect.Type]*Schema{},
		typeNames:       map[reflect.Type]string{},
		schemasByID:     map[string]*Schema{},
		mappers:         map[string][]Mapper{},
		embedded:        map[string]*Schema{},
	}
}

func (s *Schemas) Init(initFunc SchemasInitFunc) *Schemas {
	return initFunc(s)
}

func (s *Schemas) Err() error {
	return NewErrors(s.errors...)
}

func (s *Schemas) AddSchemas(schema *Schemas) *Schemas {
	for _, schema := range schema.Schemas() {
		s.AddSchema(*schema)
	}
	return s
}

func (s *Schemas) RemoveSchema(schema Schema) *Schemas {
	s.Lock()
	defer s.Unlock()
	return s.doRemoveSchema(schema)
}

func (s *Schemas) doRemoveSchema(schema Schema) *Schemas {
	delete(s.schemasByID, schema.ID)
	return s
}

func (s *Schemas) AddSchema(schema Schema) *Schemas {
	s.Lock()
	defer s.Unlock()
	return s.doAddSchema(schema)
}

func (s *Schemas) doAddSchema(schema Schema) *Schemas {
	s.setupDefaults(&schema)

	_, ok := s.schemasByID[schema.ID]
	if !ok {
		s.schemasByID[schema.ID] = &schema
		s.schemas = append(s.schemas, &schema)
	}

	return s
}

func (s *Schemas) setupDefaults(schema *Schema) {
	schema.Type = "schema"
	if schema.ID == "" {
		s.errors = append(s.errors, fmt.Errorf("ID is not set on schema: %v", schema))
		return
	}
	if schema.PluralName == "" {
		schema.PluralName = name.GuessPluralName(schema.ID)
	}
	if schema.CodeName == "" {
		schema.CodeName = convert2.Capitalize(schema.ID)
	}
	if schema.CodeNamePlural == "" {
		schema.CodeNamePlural = name.GuessPluralName(schema.CodeName)
	}
}

func (s *Schemas) AddMapper(schemaID string, mapper Mapper) *Schemas {
	s.mappers[schemaID] = append(s.mappers[schemaID], mapper)
	return s
}

func (s *Schemas) Schemas() []*Schema {
	return s.schemas
}

func (s *Schemas) SchemasByID() map[string]*Schema {
	return s.schemasByID
}

func (s *Schemas) mapper(schemaID string) []Mapper {
	return s.mappers[schemaID]
}

func (s *Schemas) Schema(name string) *Schema {
	return s.doSchema(name, true)
}

func (s *Schemas) doSchema(name string, lock bool) *Schema {
	if lock {
		s.Lock()
	}
	schema, ok := s.schemasByID[name]
	if lock {
		s.Unlock()
	}
	if ok {
		return schema
	}

	for _, check := range s.schemas {
		if strings.EqualFold(check.ID, name) || strings.EqualFold(check.PluralName, name) {
			return check
		}
	}

	return nil
}

type MultiErrors struct {
	Errors []error
}

type Errors struct {
	errors []error
}

func (e *Errors) Add(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}

func (e *Errors) Err() error {
	return NewErrors(e.errors...)
}

func NewErrors(inErrors ...error) error {
	var errors []error
	for _, err := range inErrors {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) == 0 {
		return nil
	} else if len(errors) == 1 {
		return errors[0]
	}
	return &MultiErrors{
		Errors: errors,
	}
}

func (m *MultiErrors) Error() string {
	buf := bytes.NewBuffer(nil)
	for _, err := range m.Errors {
		if buf.Len() > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(err.Error())
	}

	return buf.String()
}
