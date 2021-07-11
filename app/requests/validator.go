package requests

import (
	"gopkg.in/go-playground/validator.v9"
	"regexp"
)

var (
	slugRegex = regexp.MustCompile(`^\w+$`)
)

func Validate(request interface{}) []string {
	v := validator.New()
	_ = v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return slugRegex.MatchString(fl.Field().String())
	})
	err := v.Struct(request)

	var validationErrors []string
	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, e.Translate(nil))
		}
	}

	return validationErrors
}
