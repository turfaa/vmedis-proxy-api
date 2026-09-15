package gin2

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
)

// JSONV2Renderer is used to render data with "encoding/json/v2".
type JSONV2Renderer struct {
	Data any
}

func (r JSONV2Renderer) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	jsonBytes, err := json.Marshal(r.Data)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	_, err = w.Write(jsonBytes)
	return err
}

func (r JSONV2Renderer) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
}
