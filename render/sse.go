package render

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gowool/wool/internal"
	"net/http"
	"reflect"
)

// https://html.spec.whatwg.org/multipage/server-sent-events.html
var (
	_newLine    = []byte{10}
	_twoNewLine = []byte{10, 10}
	_id         = []byte{105, 100, 58}
	_event      = []byte{101, 118, 101, 110, 116, 58}
	_retry      = []byte{114, 101, 116, 114, 121, 58}
	_data       = []byte{100, 97, 116, 97, 58}
)

type SSEvent struct {
	Id    string
	Event string
	Retry uint
	Data  any
}

func (r SSEvent) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	if err := r.writeId(w); err != nil {
		return err
	}
	if err := r.writeEvent(w); err != nil {
		return err
	}
	if err := r.writeRetry(w); err != nil {
		return err
	}
	return r.writeData(w)
}

func (r SSEvent) WriteContentType(w http.ResponseWriter) {
	w.Header().Set(headerContentType, mimeTextEventStreamCharsetUTF8)
	w.Header().Set(headerCacheControl, "no-cache")
	w.Header().Set(headerConnection, "keep-alive")
	// https://github.com/pocketbase/pocketbase/discussions/480#discussioncomment-3657640
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering
	w.Header().Set(headerXAccelBuffering, "no")
}

func (r SSEvent) writeId(w http.ResponseWriter) error {
	return write(w, _id, internal.StringToBytes(r.Id), _newLine)
}

func (r SSEvent) writeEvent(w http.ResponseWriter) error {
	return write(w, _event, internal.StringToBytes(r.Event), _newLine)
}

func (r SSEvent) writeRetry(w http.ResponseWriter) error {
	if r.Retry > 0 {
		return write(w, _retry, internal.StringToBytes(fmt.Sprintf("%d", r.Retry)), _newLine)
	}
	return nil
}

func (r SSEvent) writeData(w http.ResponseWriter) (err error) {
	var (
		d  []byte
		ok bool
	)
	if d, ok = r.Data.([]byte); !ok {
		switch kindOfData(r.Data) {
		case reflect.Struct, reflect.Slice, reflect.Map:
			d, err = json.Marshal(r.Data)
			if err != nil {
				return
			}
		default:
			d = internal.StringToBytes(fmt.Sprint(r.Data))
		}
	}
	return write(w, _data, d, _twoNewLine)
}

func write(w http.ResponseWriter, data ...[]byte) error {
	for _, item := range data {
		if _, err := w.Write(item); err != nil {
			return err
		}
	}
	return nil
}

func kindOfData(data any) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}
