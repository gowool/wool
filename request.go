package wool

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

const defaultMaxMemory = 32 << 20 // 32 MB

type Request struct {
	*http.Request
	query       url.Values
	accept      []string
	contentType string
}

func (r *Request) WithContext(ctx context.Context) *Request {
	r2 := new(Request)
	*r2 = *r
	r2.Request = r.Request.WithContext(ctx)
	return r2
}

func (r *Request) Clone(ctx context.Context) *Request {
	r2 := new(Request)
	*r2 = *r
	r2.Request = r.Request.Clone(ctx)
	return r2
}

func (r *Request) ContentType() string {
	if r.contentType == "" {
		r.contentType = r.Header.Get(HeaderContentType)
		for i, char := range r.contentType {
			if char == ' ' || char == ';' {
				r.contentType = r.contentType[:i]
				break
			}
		}
	}
	return r.contentType
}

func (r *Request) Accept() []string {
	if r.accept == nil {
		parts := strings.Split(r.Header.Get(HeaderAccept), ",")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			if i := strings.IndexByte(part, ';'); i > 0 {
				part = part[:i]
			}
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, part)
			}
		}
		r.accept = out
	}
	return r.accept
}

func (r *Request) IsJSON() bool {
	return r.ContentType() == MIMEApplicationJSON
}

func (r *Request) IsForm() bool {
	return r.ContentType() == MIMEApplicationForm
}

func (r *Request) IsMultipartForm() bool {
	return r.ContentType() == MIMEMultipartForm
}

func (r *Request) IsTLS() bool {
	return r.TLS != nil
}

func (r *Request) PathParams() PathParams {
	return paramsFromContext(r.Context())
}

func (r *Request) PathParam(param string) string {
	if s, ok := r.PathParams()[param]; ok && len(s) > 0 {
		return s[0]
	}
	return ""
}

func (r *Request) PathParamID() string {
	return r.PathParam("id")
}

func (r *Request) QueryParams() url.Values {
	if r.query == nil {
		r.query = r.URL.Query()
	}
	return r.query
}

func (r *Request) QueryParam(name string) string {
	return r.QueryParams().Get(name)
}

func (r *Request) FormValues() (url.Values, error) {
	if r.IsMultipartForm() {
		if err := r.ParseMultipartForm(defaultMaxMemory); err != nil {
			return nil, err
		}
	} else if err := r.ParseForm(); err != nil {
		return nil, err
	}
	return r.Form, nil
}
