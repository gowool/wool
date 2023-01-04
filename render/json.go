package render

import (
	"github.com/goccy/go-json"
	"net/http"
)

type JSON struct {
	Data any
}

type IndentedJSON struct {
	Data any
}

func (r JSON) Render(w http.ResponseWriter) error {
	data, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	r.WriteContentType(w)
	_, err = w.Write(data)
	return err
}

func (r JSON) WriteContentType(w http.ResponseWriter) {
	w.Header().Set(headerContentType, mimeApplicationJSONCharsetUTF8)
}

func (r IndentedJSON) Render(w http.ResponseWriter) error {
	data, err := json.MarshalIndent(r.Data, "", "    ")
	if err != nil {
		return err
	}
	r.WriteContentType(w)
	_, err = w.Write(data)
	return err
}

func (r IndentedJSON) WriteContentType(w http.ResponseWriter) {
	w.Header().Set(headerContentType, mimeApplicationJSONCharsetUTF8)
}
