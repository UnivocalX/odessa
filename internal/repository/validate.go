package repository

import (
	"net/url"

	"example.com/aether/internal/storage"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	validate.RegisterValidation("storage_uri", func(fl validator.FieldLevel) bool {
		uri, ok := fl.Field().Interface().(storage.URI)
		if !ok {
			return false
		}

		u, err := url.Parse(string(uri))
		if err != nil {
			return false
		}

		return u.Scheme != ""
	})
}