// Package docs embeds the OpenAPI specification and the Redoc page that
// renders it, so the API documentation ships inside the binary.
package docs

import (
	_ "embed"
	"net/http"
)

// OpenAPISpec is the OpenAPI specification of the HTTP API.
//
//go:embed openapi.yaml
var OpenAPISpec []byte

// RedocHTML is the Redoc page that renders OpenAPISpec from /docs/openapi.yaml.
//
//go:embed index.html
var RedocHTML []byte

// SpecHandler serves OpenAPISpec.
func SpecHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=300, must-revalidate")
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write(OpenAPISpec)
}

// RedocHandler serves RedocHTML.
func RedocHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=3600, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(RedocHTML)
}
