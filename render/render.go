package render

import "net/http"

var (
	_ HTMLRender = (*HTMLEngine)(nil)
	_ Render     = (*HTML)(nil)
	_ Render     = (*Blob)(nil)
	_ Render     = (*JSON)(nil)
	_ Render     = (*IndentedJSON)(nil)
	_ Render     = (*String)(nil)
	_ Render     = (*Redirect)(nil)
	_ Render     = (*SSEvent)(nil)
)

const (
	headerContentType     = "Content-Type"
	headerCacheControl    = "Cache-Control"
	headerConnection      = "Connection"
	headerXAccelBuffering = "X-Accel-Buffering"

	mimeTextHTMLCharsetUTF8        = "text/html; charset=utf-8"
	mimeTextPlainCharsetUTF8       = "text/plain; charset=utf-8"
	mimeApplicationJSONCharsetUTF8 = "application/json; charset=utf-8"
	mimeTextEventStreamCharsetUTF8 = "text/event-stream; charset=utf-8"
)

type Render interface {
	Render(w http.ResponseWriter) error
	WriteContentType(w http.ResponseWriter)
}
