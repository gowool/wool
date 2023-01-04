package render

import "net/http"

type Blob struct {
	ContentType string
	Data        []byte
}

func (r Blob) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	_, err := w.Write(r.Data)
	return err
}

func (r Blob) WriteContentType(w http.ResponseWriter) {
	w.Header().Set(headerContentType, r.ContentType)
}
