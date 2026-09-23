package controller

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// newValidator returns a validator that reports the json field name in errors.
func newValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}

		return name
	})

	return validate
}

// validationMessage renders validator failures as a single clear string,
// suitable for a codes.InvalidArgument status message.
func validationMessage(err error) string {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return "invalid request"
	}

	msgs := make([]string, 0, len(validationErrs))
	for _, fieldErr := range validationErrs {
		switch fieldErr.Tag() {
		case "required":
			msgs = append(msgs, fieldErr.Field()+" is required")
		case "max":
			msgs = append(msgs, fieldErr.Field()+" must be at most "+fieldErr.Param()+" characters")
		case "uuid":
			msgs = append(msgs, fieldErr.Field()+" must be a valid UUID")
		default:
			msgs = append(msgs, fieldErr.Field()+" is invalid")
		}
	}

	return strings.Join(msgs, "; ")
}
