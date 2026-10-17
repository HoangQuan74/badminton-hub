package validation

import "github.com/go-playground/validator/v10"

var serviceValidator = validator.New()

func Struct(value any) error {
	return serviceValidator.Struct(value)
}
