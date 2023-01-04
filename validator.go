package wool

import (
	"context"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

type FailedField struct {
	Namespace string `json:"namespace,omitempty"`
	Field     string `json:"field,omitempty"`
	Tag       string `json:"tag,omitempty"`
	Value     string `json:"value,omitempty"`
	Message   string `json:"message,omitempty"`
}

type Validator interface {
	Validate(i any) error
	ValidateCtx(ctx context.Context, i any) error
}

type wrapValidator struct {
	v *validator.Validate
}

func NewValidator() Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return ""
		}
		return name
	})

	return &wrapValidator{v: v}
}

func (v *wrapValidator) Validate(i any) error {
	return v.error(v.v.Struct(i))
}

func (v *wrapValidator) ValidateCtx(ctx context.Context, i any) error {
	return v.error(v.v.StructCtx(ctx, i))
}

func (v *wrapValidator) error(err error) error {
	if err == nil {
		return nil
	}

	var data []FailedField
	for _, ve := range err.(validator.ValidationErrors) {
		data = append(data, FailedField{
			Namespace: ve.StructNamespace(),
			Field:     ve.Field(),
			Tag:       ve.Tag(),
			Value:     ve.Param(),
			Message:   ve.Error(),
		})
	}

	return NewErrUnprocessableEntity(nil, data)
}
