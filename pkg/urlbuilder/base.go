package urlbuilder

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func ParseRequestURL(r *http.Request) string {
	// Get url from standard headers
	requestURL := getURLFromStandardHeaders(r)
	if requestURL != "" {
		return requestURL
	}

	// Use incoming url
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s%s", scheme, r.Host, r.Header.Get(PrefixHeader), r.URL.Path)
}

func getURLFromStandardHeaders(r *http.Request) string {
	xForwardedProto := getOverrideHeader(r, ForwardedProtoHeader, "")
	if xForwardedProto == "" {
		return ""
	}

	host := getOverrideHeader(r, ForwardedHostHeader, "")
	if host == "" {
		host = r.Host
	}

	if host == "" {
		return ""
	}

	port := getOverrideHeader(r, ForwardedPortHeader, "")
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

	return fmt.Sprintf("%s://%s%s%s%s", xForwardedProto, host, port, r.Header.Get(PrefixHeader), r.URL.Path)
}

func getOverrideHeader(r *http.Request, header string, defaultValue string) string {
	// Need to handle comma separated hosts in X-Forwarded-For
	value := r.Header.Get(header)
	if value != "" {
		return strings.TrimSpace(strings.Split(value, ",")[0])
	}
	return defaultValue
}

func ParseResponseURLBase(currentURL string, r *http.Request) (string, error) {
	path := r.URL.Path

	index := strings.LastIndex(currentURL, path)
	if index == -1 {
		// Fallback, if we can't find path in currentURL, then we just assume the base is the root of the web request
		u, err := url.Parse(currentURL)
		if err != nil {
			return "", err
		}

		buffer := bytes.Buffer{}
		buffer.WriteString(u.Scheme)
		buffer.WriteString("://")
		buffer.WriteString(u.Host)
		return buffer.String(), nil
	}

	return currentURL[0:index], nil
}
