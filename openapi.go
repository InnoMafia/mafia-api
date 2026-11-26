package mafiaapi

import _ "embed"

//go:embed gen/openapi/mafia/platform/v1/platform.openapi.json
var openAPI string

// OpenAPI returns the generated HTTP contract.
func OpenAPI() string {
	return openAPI
}
