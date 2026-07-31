package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/UnivocalX/odessa/internal/service"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// Decode reads the JSON request body into v and validates it.
func Decode[T any](r *http.Request, v T) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
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
