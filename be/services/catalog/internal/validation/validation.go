package validation

import "github.com/go-playground/validator/v10"

var serviceValidator = validator.New()

// Struct validates a Catalog use-case input without mutating it.
func Struct(value any) error {
	return serviceValidator.Struct(value)
}
