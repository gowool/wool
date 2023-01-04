package wool

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"strings"
)

var DefaultMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

var compiledRXPatterns = map[string]*regexp.Regexp{}

type PathParams map[string][]string

type ctxPathParamsKey struct{}

type route struct {
	method   string
	segments []string
	wildcard bool
	handler  Handler
}

func (wool *Wool) Add(pattern string, handler Handler, methods ...string) {
	if contains(methods, http.MethodGet) && !contains(methods, http.MethodHead) {
		methods = append(methods, http.MethodHead)
	}

	if len(methods) == 0 {
		methods = DefaultMethods
	}

	pattern = wool.prefix + pattern

	for _, method := range methods {
		route := route{
			method:   strings.ToUpper(method),
			segments: strings.Split(pattern, "/"),
			wildcard: strings.HasSuffix(pattern, "/..."),
			handler:  wool.wrap(handler),
		}

		*wool.routes = append(*wool.routes, route)
	}

	for _, segment := range strings.Split(pattern, "/") {
		if strings.HasPrefix(segment, ":") {
			_, rxPattern, containsRx := strings.Cut(segment, "|")
			if containsRx {
				compiledRXPatterns[rxPattern] = regexp.MustCompile(rxPattern)
			}
		}
	}

	if wool.Log != nil {
		wool.Log.Info("handler registered", zap.String("pattern", pattern), zap.Strings("methods", methods))
	}
}

func paramsFromContext(ctx context.Context) PathParams {
	if params, ok := ctx.Value(ctxPathParamsKey{}).(PathParams); ok {
		return params
	}
	return PathParams{}
}

func contextWithParams(ctx context.Context, params PathParams) context.Context {
	return context.WithValue(ctx, ctxPathParamsKey{}, params)
}

func (r *route) match(ctx context.Context, urlSegments []string) (context.Context, bool) {
	if !r.wildcard && len(urlSegments) != len(r.segments) {
		return ctx, false
	}

	params := paramsFromContext(ctx)

	for i, routeSegment := range r.segments {
		if i > len(urlSegments)-1 {
			return ctx, false
		}

		if routeSegment == "..." {
			params["..."] = []string{strings.Join(urlSegments[i:], "/")}
			return contextWithParams(ctx, params), true
		}

		if routeSegment != "" && routeSegment[0] == ':' {
			key, rxPattern, containsRx := strings.Cut(routeSegment[1:], "|")

			if (containsRx && compiledRXPatterns[rxPattern].MatchString(urlSegments[i])) ||
				(!containsRx && urlSegments[i] != "") {
				params[key] = append(params[key], urlSegments[i])
				ctx = contextWithParams(ctx, params)
				continue
			}

			return ctx, false
		}

		if urlSegments[i] != routeSegment {
			return ctx, false
		}
	}

	return ctx, true
}
