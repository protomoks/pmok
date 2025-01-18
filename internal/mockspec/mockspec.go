package mockspec

import (
	"net/http"
	"strings"
)

type SpecRequest struct {
	Headers     http.Header `json:"headers"`
	Method      string      `json:"method"`
	RequestPath string      `json:"path"`
}

type SpecResponse struct {
	Headers http.Header `json:"headers"`
}

type SpecBodyResponse struct {
	SpecResponse
	Body map[string]any `json:"body"`
}

type MockWriter interface {
	WriteResponse(res *http.Response) error
	Close() error
}

func MockFileNameFromPath(p string) string {
	return strings.ReplaceAll(p, "/", "_") + ".json"
}
