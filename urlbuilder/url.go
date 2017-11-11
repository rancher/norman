package urlbuilder

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rancher/norman/name"
	"github.com/rancher/norman/types"
)

const (
	DEFAULT_OVERRIDE_URL_HEADER = "X-API-request-url"
	FORWARDED_HOST_HEADER       = "X-Forwarded-Host"
	FORWARDED_PROTO_HEADER      = "X-Forwarded-Proto"
	FORWARDED_PORT_HEADER       = "X-Forwarded-Port"
)

func New(r *http.Request, version types.APIVersion, schemas *types.Schemas) (types.URLBuilder, error) {
	requestUrl := parseRequestUrl(r)
	responseUrlBase, err := parseResponseUrlBase(requestUrl, r)
	if err != nil {
		return nil, err
	}

	builder := &urlBuilder{
		schemas:         schemas,
		requestUrl:      requestUrl,
		responseUrlBase: responseUrlBase,
		apiVersion:      version,
		query:           r.URL.Query(),
	}

	return builder, nil
}

type urlBuilder struct {
	schemas         *types.Schemas
	requestUrl      string
	responseUrlBase string
	apiVersion      types.APIVersion
	subContext      string
	query           url.Values
}

func (u *urlBuilder) SetSubContext(subContext string) {
	u.subContext = subContext
}

func (u *urlBuilder) ResourceLink(resource *types.RawResource) string {
	if resource.ID == "" {
		return ""
	}

	return u.constructBasicUrl(resource.Schema.Version, resource.Schema.PluralName, resource.ID)
}

func (u *urlBuilder) ReverseSort(order types.SortOrder) string {
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

	return u.requestUrl + "?" + newValues.Encode()
}

func (u *urlBuilder) Current() string {
	return u.requestUrl
}

func (u *urlBuilder) RelativeToRoot(path string) string {
	return u.responseUrlBase + path
}

func (u *urlBuilder) Collection(schema *types.Schema) string {
	plural := u.getPluralName(schema)
	return u.constructBasicUrl(schema.Version, plural)
}

func (u *urlBuilder) Version(version string) string {
	return fmt.Sprintf("%s/%s", u.responseUrlBase, version)
}

func (u *urlBuilder) constructBasicUrl(version types.APIVersion, parts ...string) string {
	buffer := bytes.Buffer{}

	buffer.WriteString(u.responseUrlBase)
	if version.Path == "" {
		buffer.WriteString(u.apiVersion.Path)
	} else {
		buffer.WriteString(version.Path)
	}
	buffer.WriteString(u.subContext)

	for _, part := range parts {
		if part == "" {
			return ""
		}
		buffer.WriteString("/")
		buffer.WriteString(part)
	}

	return buffer.String()
}

func (u *urlBuilder) getPluralName(schema *types.Schema) string {
	if schema.PluralName == "" {
		return strings.ToLower(name.GuessPluralName(schema.ID))
	}
	return strings.ToLower(schema.PluralName)
}

// Constructs the request URL based off of standard headers in the request, falling back to the HttpServletRequest.getRequestURL()
// if the headers aren't available. Here is the ordered list of how we'll attempt to construct the URL:
//  - x-api-request-url
//  - x-forwarded-proto://x-forwarded-host:x-forwarded-port/HttpServletRequest.getRequestURI()
//  - x-forwarded-proto://x-forwarded-host/HttpServletRequest.getRequestURI()
//  - x-forwarded-proto://host:x-forwarded-port/HttpServletRequest.getRequestURI()
//  - x-forwarded-proto://host/HttpServletRequest.getRequestURI() request.getRequestURL()
//
// Additional notes:
//  - With x-api-request-url, the query string is passed, it will be dropped to match the other formats.
//  - If the x-forwarded-host/host header has a port and x-forwarded-port has been passed, x-forwarded-port will be used.
func parseRequestUrl(r *http.Request) string {
	// Get url from custom x-api-request-url header
	requestUrl := getOverrideHeader(r, DEFAULT_OVERRIDE_URL_HEADER, "")
	if requestUrl != "" {
		return strings.SplitN(requestUrl, "?", 2)[0]
	}

	// Get url from standard headers
	requestUrl = getUrlFromStandardHeaders(r)
	if requestUrl != "" {
		return requestUrl
	}

	// Use incoming url
	return fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
}

func getUrlFromStandardHeaders(r *http.Request) string {
	xForwardedProto := getOverrideHeader(r, FORWARDED_PROTO_HEADER, "")
	if xForwardedProto == "" {
		return ""
	}

	host := getOverrideHeader(r, FORWARDED_HOST_HEADER, "")
	if host == "" {
		host = r.Host
	}

	if host == "" {
		return ""
	}

	port := getOverrideHeader(r, FORWARDED_PORT_HEADER, "")
	if port == "443" || port == "80" {
		port = "" // Don't include default ports in url
	}

	if port != "" && strings.Contains(host, ":") {
		// Have to strip the port that is in the host. Handle IPv6, which has this format: [::1]:8080
		if (strings.HasPrefix(host, "[") && strings.Contains(host, "]:")) || !strings.HasPrefix(host, "[") {
			host = host[0:strings.LastIndex(host, ":")]
		}
	}

	if port != "" {
		port = ":" + port
	}

	return fmt.Sprintf("%s://%s%s%s", xForwardedProto, host, port, r.URL.Path)
}

func getOverrideHeader(r *http.Request, header string, defaultValue string) string {
	// Need to handle comma separated hosts in X-Forwarded-For
	value := r.Header.Get(header)
	if value != "" {
		return strings.TrimSpace(strings.Split(value, ",")[0])
	}
	return defaultValue
}

func parseResponseUrlBase(requestUrl string, r *http.Request) (string, error) {
	path := r.URL.Path

	index := strings.LastIndex(requestUrl, path)
	if index == -1 {
		// Fallback, if we can't find path in requestUrl, then we just assume the base is the root of the web request
		u, err := url.Parse(requestUrl)
		if err != nil {
			return "", err
		}

		buffer := bytes.Buffer{}
		buffer.WriteString(u.Scheme)
		buffer.WriteString("://")
		buffer.WriteString(u.Host)
		return buffer.String(), nil
	} else {
		return requestUrl[0:index], nil
	}
}
