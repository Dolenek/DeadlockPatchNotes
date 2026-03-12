package httpapi

import _ "embed"

var (
	//go:embed static/scalar.html
	scalarHTML []byte
	//go:embed static/openapi.json
	openAPISpecJSON []byte
)
