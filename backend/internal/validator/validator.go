package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func Init() {
	validate = validator.New()
}

func Validate(s interface{}) map[string]string {
	errs := validate.Struct(s)
	if errs == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range errs.(validator.ValidationErrors) {
		field := err.Field()
		switch err.Tag() {
		case "required":
			errors[field] = fmt.Sprintf("%s is required", field)
		case "min":
			errors[field] = fmt.Sprintf("%s must be at least %s", field, err.Param())
		case "max":
			errors[field] = fmt.Sprintf("%s must be at most %s", field, err.Param())
		case "oneof":
			errors[field] = fmt.Sprintf("%s must be one of %s", field, err.Param())
		case "gt":
			errors[field] = fmt.Sprintf("%s must be greater than %s", field, err.Param())
		case "uuid":
			errors[field] = fmt.Sprintf("%s must be a valid UUID", field)
		default:
			errors[field] = fmt.Sprintf("%s is invalid", field)
		}
	}

	return errors
}

func GetFirstError(errs map[string]string) string {
	for _, err := range errs {
		return err
	}
	return "validation error"
}
