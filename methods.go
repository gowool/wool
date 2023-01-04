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

func (wool *Wool) Get(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodGet, http.MethodHead)
}

func (wool *Wool) Head(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodHead)
}

func (wool *Wool) Post(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodPost)
}

func (wool *Wool) Put(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodPut)
}

func (wool *Wool) Patch(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodPatch)
}

func (wool *Wool) Delete(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodDelete)
}

func (wool *Wool) Connect(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodConnect)
}

func (wool *Wool) Options(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodOptions)
}

func (wool *Wool) Trace(pattern string, handler Handler) {
	wool.Add(pattern, handler, http.MethodTrace)
}

func (wool *Wool) CRUD(pattern string, resource any, mw ...Middleware) {
	wool.Group(pattern, func(group *Wool) {
		group.Use(mw...)

		if resource, ok := resource.(List); ok {
			group.Get("", resource.List)
		}

		if resource, ok := resource.(Create); ok {
			group.Post("", resource.Create)
		}

		if resource, ok := resource.(Take); ok {
			group.Get(patternID, resource.Take)
		}

		if resource, ok := resource.(Update); ok {
			group.Put(patternID, resource.Update)
		}

		if resource, ok := resource.(PartiallyUpdate); ok {
			group.Patch(patternID, resource.PartiallyUpdate)
		}

		if resource, ok := resource.(Delete); ok {
			group.Delete(patternID, resource.Delete)
		}
	})
}

func (wool *Wool) MountHealth() {
	wool.Get("/health", func(c Ctx) error {
		return c.NoContent()
	})
}
