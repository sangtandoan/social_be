package validator

import "github.com/go-playground/validator/v10"

type CustomValidator struct {
	validator *validator.Validate
}

func New() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

func (v *CustomValidator) Validate(data any) error {
	return v.validator.Struct(data)
}
