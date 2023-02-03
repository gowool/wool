package wool

import (
	"errors"
	"fmt"
	"github.com/gowool/wool/render"
	"golang.org/x/exp/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	DefaultNotFoundHandler = func(c Ctx) error {
		return NewErrNotFound(nil)
	}

	DefaultMethodNotAllowed = func(Ctx) error {
		return NewErrMethodNotAllowed(nil)
	}

	DefaultOptionsHandler = func(c Ctx) error {
		return c.NoContent()
	}

	DefaultErrorHandler = func(c Ctx, err *Error) error {
		switch c.NegotiateFormat(MIMETextPlain, MIMETextHTML, MIMEApplicationJSON) {
		case MIMEApplicationJSON:
			return c.JSON(err.Code, err)
		default:
			if c.Debug() {
				return c.String(err.Code, "code=%d, message=%v, data=%v, developer_message=%s", err.Code, err.Message, err.Data, err.Developer)
			}
			return c.String(err.Code, "code=%d, message=%v, data=%v", err.Code, err.Message, err.Data)
		}
	}

	DefaultErrorTransform = func(err error) *Error {
		var e *Error
		if !errors.As(err, &e) {
			e = NewError(http.StatusInternalServerError, err)
		}
		return e
	}
)

type (
	Handler        func(c Ctx) error
	Middleware     func(next Handler) Handler
	ErrorHandler   func(c Ctx, err *Error) error
	ErrorTransform func(err error) *Error
	AfterServe     func(c Ctx, start, end time.Time, err error)
)

type Wool struct {
	Debug            bool
	NewCtxFunc       func(wool *Wool, r *http.Request, w http.ResponseWriter) Ctx
	HTMLRender       render.HTMLRender
	NotFoundHandler  Handler
	MethodNotAllowed Handler
	OptionsHandler   Handler
	ErrorHandler     ErrorHandler
	ErrorTransform   ErrorTransform
	AfterServe       AfterServe
	Validator        Validator
	middlewares      []Middleware
	ctxPool          *sync.Pool
	routes           *[]route
	prefix           string
}

func ToHandler(handler http.Handler) Handler {
	return func(c Ctx) error {
		handler.ServeHTTP(c.Res(), c.Req().Request)
		return nil
	}
}

func ToMiddleware(wrapper func(http.Handler) http.Handler) Middleware {
	return func(next Handler) Handler {
		return func(c Ctx) (err error) {
			wrapper(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.SetReq(c.Req().WithContext(r.Context()))
				err = next(c)
			})).ServeHTTP(c.Res(), c.Req().Request)
			return
		}
	}
}

type Option func(*Wool)

func WithNewCtxFunc(newCtxFunc func(*Wool, *http.Request, http.ResponseWriter) Ctx) Option {
	return func(w *Wool) {
		w.NewCtxFunc = newCtxFunc
	}
}

func WithHTMLRender(r render.HTMLRender) Option {
	return func(w *Wool) {
		w.HTMLRender = r
	}
}

func WithNotFoundHandler(h Handler) Option {
	return func(w *Wool) {
		w.NotFoundHandler = h
	}
}

func WithMethodNotAllowed(h Handler) Option {
	return func(w *Wool) {
		w.MethodNotAllowed = h
	}
}

func WithOptionsHandler(h Handler) Option {
	return func(w *Wool) {
		w.OptionsHandler = h
	}
}

func WithErrorHandler(h ErrorHandler) Option {
	return func(w *Wool) {
		w.ErrorHandler = h
	}
}

func WithErrorTransform(et ErrorTransform) Option {
	return func(w *Wool) {
		w.ErrorTransform = et
	}
}

func WithAfterServe(as AfterServe) Option {
	return func(w *Wool) {
		w.AfterServe = as
	}
}

func WithValidator(v Validator) Option {
	return func(w *Wool) {
		w.Validator = v
	}
}

func WithMiddleware(mw ...Middleware) Option {
	return func(w *Wool) {
		w.Use(mw...)
	}
}

func New(options ...Option) *Wool {
	wool := &Wool{
		Debug:            Logger().Enabled(slog.LevelDebug),
		NewCtxFunc:       NewCtx,
		HTMLRender:       &render.HTMLEngine{},
		NotFoundHandler:  DefaultNotFoundHandler,
		MethodNotAllowed: DefaultMethodNotAllowed,
		OptionsHandler:   DefaultOptionsHandler,
		ErrorHandler:     DefaultErrorHandler,
		ErrorTransform:   DefaultErrorTransform,
		Validator:        NewValidator(),
		ctxPool:          &sync.Pool{},
		routes:           &[]route{},
	}
	wool.ctxPool.New = func() any {
		return wool.NewCtx(nil, nil)
	}
	for _, opt := range options {
		opt(wool)
	}
	return wool
}

func (wool *Wool) NewCtx(r *http.Request, w http.ResponseWriter) Ctx {
	return wool.NewCtxFunc(wool, r, w)
}

func (wool *Wool) Use(mw ...Middleware) {
	wool.middlewares = append(wool.middlewares, mw...)
}

func (wool *Wool) Group(pattern string, fn func(*Wool)) {
	mm := *wool
	mm.prefix += pattern
	fn(&mm)
}

func (wool *Wool) AcquireCtx() Ctx {
	return wool.ctxPool.Get().(Ctx)
}

func (wool *Wool) ReleaseCtx(c Ctx) {
	wool.ctxPool.Put(c)
}

func (wool *Wool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := wool.AcquireCtx()
	c.Reset(r, w)

	var start, end time.Time

	if wool.AfterServe != nil {
		start = time.Now()
	}

	err := wool.serve(c)

	if wool.AfterServe != nil {
		end = time.Now()
		wool.AfterServe(c, start, end, err)
	}

	wool.ReleaseCtx(c)
}

func (wool *Wool) Error(next Handler) Handler {
	return func(c Ctx) error {
		if err := next(c); err != nil {
			e := wool.ErrorTransform(err)

			if c.Debug() && e.Internal != nil {
				e.Developer = e.Internal.Error()
			}

			if err = wool.ErrorHandler(c, e); err != nil {
				Logger().Error("UNKNOWN ERROR", err)
			}

			return e
		}
		return nil
	}
}

func (wool *Wool) Recover(next Handler) Handler {
	return func(c Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				if err, ok = r.(error); !ok {
					err = fmt.Errorf("%v", r)
				}

				var brokenPipe bool
				if ne, ok := r.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Req().Request, false)
				if brokenPipe {
					Logger().Error(c.Req().URL.Path, err, "request", string(httpRequest))
					return
				}

				stack := make([]byte, 4<<10) // 4KB
				length := runtime.Stack(stack, true)
				stack = stack[:length]

				Logger().Error("recover from panic", err, "request", string(httpRequest), "stack", string(stack))
			}
		}()

		return next(c)
	}
}

func (wool *Wool) serve(c Ctx) error {
	urlSegments := strings.Split(c.Req().URL.Path, "/")
	allowedMethods := make([]string, 0, len(DefaultMethods))

	for _, route := range *wool.routes {
		ctx, ok := route.match(c.Req().Context(), urlSegments)
		if ok {
			if c.Req().Method == route.method {
				c.SetReq(c.Req().WithContext(ctx))
				return route.handler(c)
			}
			if !contains(allowedMethods, route.method) {
				allowedMethods = append(allowedMethods, route.method)
			}
		}
	}

	if len(allowedMethods) > 0 {
		c.Res().Header().Set("Allow", strings.Join(append(allowedMethods, http.MethodOptions), ","))
		if c.Req().Method == http.MethodOptions {
			return wool.wrap(wool.OptionsHandler)(c)
		}
		return wool.wrap(wool.MethodNotAllowed)(c)
	}

	return wool.wrap(wool.NotFoundHandler)(c)
}

func (wool *Wool) wrap(handler Handler) Handler {
	for i := len(wool.middlewares) - 1; i >= 0; i-- {
		handler = wool.middlewares[i](handler)
	}

	handler = wool.Recover(handler)
	handler = wool.Error(handler)

	return handler
}
