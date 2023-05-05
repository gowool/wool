package wool

import (
	"errors"
	"github.com/gowool/wool/render"
	"io"
	"net/http"
)

var _ CtxRender = (*DefaultCtx)(nil)

var ErrStreamClosed = errors.New("http: Stream closed")

type Map map[string]any

type CtxRender interface {
	Status(status int) error
	Render(status int, r render.Render) error
	Blob(status int, contentType string, data []byte) error
	JSON(status int, obj any) error
	IndentedJSON(status int, obj any) error
	HTML(status int, name string, obj any) error
	String(status int, format string, data ...any) error
	SSEvent(event string, data any) error
	Stream(step func(w io.Writer) error) error
	Redirect(code int, location string) error
	Created(location string) error
	NoContent() error
	OK() error
}

func (c *DefaultCtx) Status(status int) error {
	c.Res().WriteHeader(status)
	c.Res().WriteHeaderNow()
	return nil
}

func (c *DefaultCtx) Render(status int, r render.Render) error {
	c.Res().WriteHeader(status)

	if !bodyAllowedForStatus(status) {
		r.WriteContentType(c.Res())
		c.Res().WriteHeaderNow()
		return nil
	}
	return r.Render(c.Res())
}

func (c *DefaultCtx) Blob(status int, contentType string, data []byte) error {
	return c.Render(status, render.Blob{ContentType: contentType, Data: data})
}

func (c *DefaultCtx) JSON(status int, obj any) error {
	return c.Render(status, render.JSON{Data: obj})
}

func (c *DefaultCtx) IndentedJSON(status int, obj any) error {
	return c.Render(status, render.IndentedJSON{Data: obj})
}

func (c *DefaultCtx) HTML(status int, name string, obj any) error {
	instance := c.wool.HTMLRender.Instance(name, obj, c.Debug())
	return c.Render(status, instance)
}

func (c *DefaultCtx) String(status int, format string, data ...any) error {
	return c.Render(status, render.String{Format: format, Data: data})
}

func (c *DefaultCtx) SSEvent(event string, data any) error {
	return c.Render(-1, render.SSEvent{Event: event, Data: data})
}

func (c *DefaultCtx) Stream(step func(w io.Writer) error) error {
	for {
		select {
		case <-c.Req().Context().Done():
			return nil
		default:
			err := step(c.Res())
			switch err {
			case nil:
				c.Res().Flush()
			case ErrStreamClosed:
				c.Res().Flush()
				return nil
			default:
				c.wool.Log.Error("stream error", "err", err)
				return err
			}
		}
	}
}

func (c *DefaultCtx) Redirect(code int, location string) error {
	return c.Render(-1, render.Redirect{Code: code, Location: location, Request: c.Req().Request})
}

func (c *DefaultCtx) Created(location string) error {
	if location == "" {
		return c.Status(http.StatusCreated)
	}
	return c.Redirect(http.StatusCreated, location)
}

func (c *DefaultCtx) NoContent() error {
	return c.Status(http.StatusNoContent)
}

func (c *DefaultCtx) OK() error {
	return c.Status(http.StatusOK)
}

func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}
