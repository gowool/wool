package wool

import (
	"bufio"
	"fmt"
	"github.com/gowool/wool/internal"
	"net"
	"net/http"
)

var _ Response = (*response)(nil)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

type Response interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	Pusher() http.Pusher
	Status() int
	Size() int64
	Written() bool
	WriteString(s string) (int, error)
	WriteHeaderNow()
}

type response struct {
	http.ResponseWriter
	status int
	size   int64
}

func NewResponse(w http.ResponseWriter) Response {
	return &response{ResponseWriter: w, status: defaultStatus, size: noWritten}
}

func (r *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if r.size < 0 {
		r.size = 0
	}
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

func (r *response) Flush() {
	r.WriteHeaderNow()
	r.ResponseWriter.(http.Flusher).Flush()
}

func (r *response) Pusher() http.Pusher {
	if pusher, ok := r.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}

func (r *response) Status() int {
	return r.status
}

func (r *response) Size() int64 {
	return r.size
}

func (r *response) Written() bool {
	return r.size != noWritten
}

func (r *response) WriteHeader(status int) {
	if status > 0 && r.status != status {
		if r.Written() {
			Logger().Warn(fmt.Sprintf("Headers were already written. Wanted to override status code %d with %d\n", r.status, status))
			return
		}
		r.status = status
	}
}

func (r *response) Write(data []byte) (n int, err error) {
	r.WriteHeaderNow()
	n, err = r.ResponseWriter.Write(data)
	r.size += int64(n)
	return
}

func (r *response) WriteString(s string) (int, error) {
	return r.Write(internal.StringToBytes(s))
}

func (r *response) WriteHeaderNow() {
	if !r.Written() {
		r.size = 0
		r.ResponseWriter.WriteHeader(r.status)
	}
}
