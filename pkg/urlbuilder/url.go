package urlbuilder

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/rancher/norman/pkg/types"
	"github.com/rancher/wrangler/pkg/name"
)

const (
	PrefixHeader         = "X-API-URL-Prefix"
	ForwardedHostHeader  = "X-Forwarded-Host"
	ForwardedProtoHeader = "X-Forwarded-Proto"
	ForwardedPortHeader  = "X-Forwarded-Port"
)

func New(r *http.Request, resolver PathResolver, schemas *types.Schemas) (types.URLBuilder, error) {
	requestURL := ParseRequestURL(r)
	responseURLBase, err := ParseResponseURLBase(requestURL, r)
	if err != nil {
		return nil, err
	}

	builder := &DefaultURLBuilder{
		schemas:         schemas,
		currentURL:      requestURL,
		responseURLBase: responseURLBase,
		pathResolver:    resolver,
		query:           r.URL.Query(),
	}

	return builder, nil
}

type PathResolver interface {
	Schema(base string, schema *types.Schema) string
}

type DefaultPathResolver struct {
	Schemas *types.Schemas
}

func (d *DefaultPathResolver) Schema(base string, schema *types.Schema) string {
	if schema.ID == "schema" {
		return constructBasicURL(base, "schemas")
	}
	return constructBasicURL(base, schema.PluralName)
}

type DefaultURLBuilder struct {
	pathResolver    PathResolver
	schemas         *types.Schemas
	currentURL      string
	responseURLBase string
	query           url.Values
}

func (u *DefaultURLBuilder) Link(schema *types.Schema, id string, linkName string) string {
	return u.schemaURL(schema, id, linkName)
}

func (u *DefaultURLBuilder) ResourceLink(schema *types.Schema, id string) string {
	return u.schemaURL(schema, id)
}

func (u *DefaultURLBuilder) Marker(marker string) string {
	newValues := url.Values{}
	for k, v := range u.query {
		newValues[k] = v
	}
	newValues.Set("marker", marker)
	return u.Current() + "?" + newValues.Encode()
}

func (u *DefaultURLBuilder) ReverseSort(order types.SortOrder) string {
	newValues := url.Values{}
	for k, v := range u.query {
		newValues[k] = v
	}
	newValues.Del("order")
	newValues.Del("marker")
	if order == types.ASC {
		newValues.Add("order", string(types.DESC))
	} else {
		newValues.Add("order", string(types.ASC))
	}

	return u.Current() + "?" + newValues.Encode()
}

func (u *DefaultURLBuilder) Current() string {
	return u.currentURL
}

func (u *DefaultURLBuilder) RelativeToRoot(path string) string {
	return u.responseURLBase + path
}

func (u *DefaultURLBuilder) Sort(field string) string {
	newValues := url.Values{}
	for k, v := range u.query {
		newValues[k] = v
	}
	newValues.Del("order")
	newValues.Del("marker")
	newValues.Set("sort", field)
	return u.Current() + "?" + newValues.Encode()
}

func (u *DefaultURLBuilder) Collection(schema *types.Schema) string {
	return u.schemaURL(schema)
}

func (u *DefaultURLBuilder) FilterLink(schema *types.Schema, fieldName string, value string) string {
	return u.schemaURL(schema) + "?" +
		url.QueryEscape(fieldName) + "=" + url.QueryEscape(value)
}

func (u *DefaultURLBuilder) schemaURL(schema *types.Schema, parts ...string) string {
	base := []string{
		u.pathResolver.Schema(u.responseURLBase, schema),
	}
	return constructBasicURL(append(base, parts...)...)
}

func constructBasicURL(parts ...string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	default:
		base := parts[0]
		rest := path.Join(parts[1:]...)
		if !strings.HasSuffix(base, "/") && !strings.HasPrefix(rest, "/") {
			return base + "/" + rest
		}
		return base + rest
	}
}

func (u *DefaultURLBuilder) getPluralName(schema *types.Schema) string {
	if schema.PluralName == "" {
		return strings.ToLower(name.GuessPluralName(schema.ID))
	}
	return strings.ToLower(schema.PluralName)
}

func (u *DefaultURLBuilder) Action(schema *types.Schema, id, action string) string {
	return u.schemaURL(schema, id) + "?action=" + url.QueryEscape(action)
}

func (u *DefaultURLBuilder) CollectionAction(schema *types.Schema, action string) string {
	return u.schemaURL(schema) + "?action=" + url.QueryEscape(action)
}
