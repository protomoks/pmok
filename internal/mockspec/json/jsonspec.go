package json

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/protomoks/pmok/internal/mockspec"
)

type spec struct {
	w        io.WriteCloser
	Request  mockspec.SpecRequest      `json:"request"`
	Response mockspec.SpecBodyResponse `json:"response"`
}

func New(w io.WriteCloser) mockspec.MockWriter {
	return &spec{
		w: w,
	}
}

func (s *spec) WriteResponse(res *http.Response) error {
	s.Request.Headers = res.Header
	s.Request.RequestPath = res.Request.URL.Path
	s.Request.Method = res.Request.Method
	s.Response.Headers = res.Header

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var body map[string]any
	if err := json.Unmarshal(b, &body); err != nil {
		return err
	}

	s.Response.Body = body

	enc := json.NewEncoder(s.w)
	enc.SetIndent(" ", " ")
	return enc.Encode(s)

}

func (s *spec) Close() error {
	return s.w.Close()
}
