package wool

import "net/http"

const patternID = "/:id"

type (
	List interface {
		List(Ctx) error
	}
	Take interface {
		Take(Ctx) error
	}
	Create interface {
		Create(Ctx) error
	}
	Update interface {
		Update(Ctx) error
	}
	PartiallyUpdate interface {
		PartiallyUpdate(Ctx) error
	}
	Delete interface {
		Delete(Ctx) error
	}
)

func (wool *Wool) GET(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodGet, http.MethodHead)
}

func (wool *Wool) HEAD(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodHead)
}

func (wool *Wool) POST(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodPost)
}

func (wool *Wool) PUT(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodPut)
}

func (wool *Wool) PATCH(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodPatch)
}

func (wool *Wool) DELETE(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodDelete)
}

func (wool *Wool) CONNECT(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodConnect)
}

func (wool *Wool) OPTIONS(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodOptions)
}

func (wool *Wool) TRACE(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodTrace)
}

func (wool *Wool) CRUD(pattern string, resource any, mw ...Middleware) {
	wool.Group(pattern, func(group *Wool) {
		group.Use(mw...)

		if r, ok := resource.(List); ok {
			group.GET("", r.List)
		}

		if r, ok := resource.(Create); ok {
			group.POST("", r.Create)
		}

		if r, ok := resource.(Take); ok {
			group.GET(patternID, r.Take)
		}

		if r, ok := resource.(Update); ok {
			group.PUT(patternID, r.Update)
		}

		if r, ok := resource.(PartiallyUpdate); ok {
			group.PATCH(patternID, r.PartiallyUpdate)
		}

		if r, ok := resource.(Delete); ok {
			group.DELETE(patternID, r.Delete)
		}
	})
}

func (wool *Wool) MountHealth() {
	wool.GET("/health", func(c Ctx) error {
		return c.NoContent()
	})
}
