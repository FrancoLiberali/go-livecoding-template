package controller

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

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
		default:
			msgs = append(msgs, fieldErr.Field()+" is invalid")
		}
	}

	return strings.Join(msgs, "; ")
}

// runWithTimeout runs op under a deadline, guaranteeing a context error if the
// deadline elapses before op returns (see the HTTP controller for rationale).
func runWithTimeout[T any](
	parent context.Context,
	timeout time.Duration,
	op func(context.Context) (T, error),
) (T, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	type result struct {
		val T
		err error
	}

	done := make(chan result, 1)

	go func() {
		val, err := op(ctx)
		done <- result{val: val, err: err}
	}()

	select {
	case <-ctx.Done():
		var zero T

		return zero, ctx.Err()
	case res := <-done:
		return res.val, res.err
	}
}
