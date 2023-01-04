package render

import (
	"fmt"
	"github.com/gowool/wool/internal"
	"net/http"
)

type String struct {
	Format string
	Data   []any
}

func (r String) Render(w http.ResponseWriter) (err error) {
	r.WriteContentType(w)
	if len(r.Data) > 0 {
		_, err = fmt.Fprintf(w, r.Format, r.Data...)
		return
	}
	_, err = w.Write(internal.StringToBytes(r.Format))
	return
}

func (r String) WriteContentType(w http.ResponseWriter) {
	w.Header().Set(headerContentType, mimeTextPlainCharsetUTF8)
}
