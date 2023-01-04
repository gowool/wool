package wool

import (
	"go.uber.org/zap"
	"net/http"
	"sync"
)

type Ctx interface {
	CtxRender
	CtxBinding
	Debug() bool
	Log() *zap.Logger
	Store() map[string]any
	Set(key string, value any)
	Get(key string) any
	Req() *Request
	SetReq(r *Request)
	Res() Response
	SetRes(r Response)
	Reset(r *http.Request, w http.ResponseWriter)
	NegotiateFormat(offered ...string) string
}

type DefaultCtx struct {
	wool  *Wool
	res   Response
	req   *Request
	store *sync.Map
}

func NewCtx(wool *Wool, r *http.Request, w http.ResponseWriter) Ctx {
	c := &DefaultCtx{wool: wool}
	c.Reset(r, w)
	return c
}

func (c *DefaultCtx) Debug() bool {
	return c.wool.Debug
}

func (c *DefaultCtx) Log() *zap.Logger {
	return c.wool.Log
}

func (c *DefaultCtx) Store() map[string]any {
	all := map[string]any{}
	if c.store != nil {
		c.store.Range(func(key, value any) bool {
			all[key.(string)] = value
			return true
		})
	}
	return all
}

func (c *DefaultCtx) Set(key string, value any) {
	if c.store == nil {
		c.store = &sync.Map{}
	}
	c.store.Store(key, value)
}

func (c *DefaultCtx) Get(key string) any {
	if c.store == nil {
		return nil
	}
	if value, ok := c.store.Load(key); ok {
		return value
	}
	return nil
}

func (c *DefaultCtx) Req() *Request {
	return c.req
}

func (c *DefaultCtx) SetReq(req *Request) {
	c.req = req
}

func (c *DefaultCtx) Res() Response {
	return c.res
}

func (c *DefaultCtx) SetRes(res Response) {
	c.res = res
}

func (c *DefaultCtx) Reset(r *http.Request, w http.ResponseWriter) {
	c.req = &Request{Request: r}
	c.res = NewResponse(w)
	c.store = nil
}

func (c *DefaultCtx) NegotiateFormat(offered ...string) string {
	if len(offered) == 0 {
		return ""
	}
	if len(c.Req().Accept()) == 0 {
		return offered[0]
	}
	for _, accepted := range c.Req().Accept() {
		for _, offer := range offered {
			i := 0
			for ; i < len(accepted); i++ {
				if accepted[i] == '*' || offer[i] == '*' {
					return offer
				}
				if accepted[i] != offer[i] {
					break
				}
			}
			if i == len(accepted) {
				return offer
			}
		}
	}
	return ""
}
