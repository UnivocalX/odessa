package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/UnivocalX/odessa/internal/service"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// Decode reads the JSON request body into v and validates it.
func Decode[T any](r *http.Request, v T) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("%w: decode request: %w", service.ErrValidation, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("%w: request must contain one JSON value", service.ErrValidation)
		}
		return fmt.Errorf("%w: trailing request data: %w", service.ErrValidation, err)
	}

	if err := validate.Struct(v); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return fmt.Errorf("%w: %w", service.ErrValidation, err)
		}
		return err
	}

	return nil
}
