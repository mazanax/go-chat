package requests

import (
	"gopkg.in/go-playground/validator.v9"
)

func Validate(request interface{}) []string {
	v := validator.New()
	err := v.Struct(request)

	var validationErrors []string
	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, e.Translate(nil))
		}
	}

	return validationErrors
}
