package wool

import (
	"encoding"
	"errors"
	"github.com/goccy/go-json"
	"github.com/spf13/cast"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

var _ CtxBinding = (*DefaultCtx)(nil)

type CtxBinding interface {
	BindBody(i any) error
	BindJSON(i any) error
	BindForm(i any) error
	BindPath(i any) error
	BindQuery(i any) error
	BindHeaders(i any) error
	BindCtx(i any) error
	Bind(i any) error
	Validate(i any) error
}

func (c *DefaultCtx) BindBody(i any) error {
	if c.Req().IsJSON() {
		return c.BindJSON(i)
	} else if c.Req().IsForm() || c.Req().IsMultipartForm() {
		return c.BindForm(i)
	}
	return NewErrBadRequest(nil)
}

func (c *DefaultCtx) BindJSON(i any) error {
	if err := json.NewDecoder(c.Req().Body).Decode(i); err != nil {
		return NewErrBadRequest(err)
	}
	return nil
}

func (c *DefaultCtx) BindForm(i any) (err error) {
	var values url.Values
	if values, err = c.Req().FormValues(); err == nil {
		if err = Bind(i, values, "form"); err == nil {
			return
		}
	}
	return NewErrBadRequest(err)
}

func (c *DefaultCtx) BindPath(i any) (err error) {
	if err = Bind(i, c.Req().PathParams(), "path"); err != nil {
		err = NewErrBadRequest(err)
	}
	return
}

func (c *DefaultCtx) BindQuery(i any) (err error) {
	if err = Bind(i, c.Req().QueryParams(), "query"); err != nil {
		err = NewErrBadRequest(err)
	}
	return
}

func (c *DefaultCtx) BindHeaders(i any) (err error) {
	if err = Bind(i, c.Req().Header, "header"); err != nil {
		err = NewErrBadRequest(err)
	}
	return
}

func (c *DefaultCtx) BindCtx(i any) error {
	if c.store != nil {
		all := url.Values{}
		c.store.Range(func(key, value any) bool {
			if v, err := cast.ToStringE(value); err == nil {
				all[key.(string)] = []string{v}
			}
			return true
		})
		if err := Bind(i, all, "ctx"); err != nil {
			return NewErrBadRequest(err)
		}
	}
	return nil
}

func (c *DefaultCtx) Bind(i any) (err error) {
	if err = c.BindPath(i); err != nil {
		return
	}
	switch c.Req().Method {
	case http.MethodGet, http.MethodHead, http.MethodDelete:
		if err = c.BindQuery(i); err != nil {
			return
		}
	}
	if err = c.BindHeaders(i); err != nil {
		return
	}
	if err = c.BindBody(i); err != nil {
		return
	}
	if err = c.BindCtx(i); err != nil {
		return err
	}

	if c.wool.Validator != nil {
		return c.Validate(i)
	}
	return nil
}

func (c *DefaultCtx) Validate(i any) error {
	if c.wool.Validator == nil {
		panic("nil validator")
	}

	return c.wool.Validator.ValidateCtx(c.Req().Context(), i)
}

func Bind(destination any, data map[string][]string, tag string) error {
	if destination == nil || len(data) == 0 {
		return nil
	}
	typ := reflect.TypeOf(destination).Elem()
	val := reflect.ValueOf(destination).Elem()

	if typ.Kind() == reflect.Map {
		for k, v := range data {
			val.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v[0]))
		}
		return nil
	}

	if typ.Kind() != reflect.Struct {
		if tag == "path" || tag == "query" || tag == "header" {
			return nil
		}
		return errors.New("binding element must be a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if typeField.Anonymous {
			if structField.Kind() == reflect.Ptr {
				structField = structField.Elem()
			}
		}
		if !structField.CanSet() {
			continue
		}
		structFieldKind := structField.Kind()
		inputFieldName := typeField.Tag.Get(tag)
		if typeField.Anonymous && structField.Kind() == reflect.Struct && inputFieldName != "" {
			return errors.New("query/path/form tags are not allowed with anonymous struct field")
		}

		if inputFieldName == "" {
			if structFieldKind == reflect.Struct {
				if err := Bind(structField.Addr().Interface(), data, tag); err != nil {
					return err
				}
			}
			continue
		}

		inputValue, exists := data[inputFieldName]
		if !exists {
			for k, v := range data {
				if strings.EqualFold(k, inputFieldName) {
					inputValue = v
					exists = true
					break
				}
			}
		}

		if !exists {
			continue
		}

		if ok, err := unmarshalField(typeField.Type.Kind(), inputValue[0], structField); ok {
			if err != nil {
				return err
			}
			continue
		}

		numElems := len(inputValue)
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for j := 0; j < numElems; j++ {
				if err := setWithProperType(sliceOf, inputValue[j], slice.Index(j)); err != nil {
					return err
				}
			}
			val.Field(i).Set(slice)
		} else if err := setWithProperType(typeField.Type.Kind(), inputValue[0], structField); err != nil {
			return err
		}
	}
	return nil
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	if ok, err := unmarshalField(valueKind, val, structField); ok {
		return err
	}

	switch valueKind {
	case reflect.Ptr:
		return setWithProperType(structField.Elem().Kind(), val, structField.Elem())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := cast.ToInt64E(val)
		if err != nil {
			return err
		}
		structField.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := cast.ToUint64E(val)
		if err != nil {
			return err
		}
		structField.SetUint(v)
	case reflect.Bool:
		v, err := cast.ToBoolE(val)
		if err != nil {
			return err
		}
		structField.SetBool(v)
	case reflect.Float32, reflect.Float64:
		v, err := cast.ToFloat64E(val)
		if err != nil {
			return err
		}
		structField.SetFloat(v)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("unknown type")
	}
	return nil
}

func unmarshalField(valueKind reflect.Kind, val string, field reflect.Value) (bool, error) {
	switch valueKind {
	case reflect.Ptr:
		return unmarshalFieldPtr(val, field)
	default:
		return unmarshalFieldNonPtr(val, field)
	}
}

func unmarshalFieldNonPtr(value string, field reflect.Value) (bool, error) {
	fieldIValue := field.Addr().Interface()
	if unmarshaler, ok := fieldIValue.(encoding.TextUnmarshaler); ok {
		return true, unmarshaler.UnmarshalText([]byte(value))
	}
	return false, nil
}

func unmarshalFieldPtr(value string, field reflect.Value) (bool, error) {
	if field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
	}
	return unmarshalFieldNonPtr(value, field.Elem())
}
