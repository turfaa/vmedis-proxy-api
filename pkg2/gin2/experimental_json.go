package gin2

import (
	"fmt"
	"net/http"

	"github.com/go-json-experiment/json"
)

// ExperimentalJSONRenderer is used to render data with the experimental "encoding/json/v2" candidate.
type ExperimentalJSONRenderer struct {
	Data any
}

func (r ExperimentalJSONRenderer) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	jsonBytes, err := json.Marshal(r.Data)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	_, err = w.Write(jsonBytes)
	return err
}

func (r ExperimentalJSONRenderer) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
}
