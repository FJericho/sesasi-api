package helper

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func GenerateValidationErrors(err error) string {
	errorMessages := map[string]string{
		"required": "is required.",
		"min":      "must be at least %s characters long.",
		"max":      "must be at most %s characters long.",
		"email":    "must be a valid email address.",
		"alphanum": "must contain only alphanumeric characters.",
		"eqfield":  "must match the field %s.",
	}

	var messages []string
	for _, validationErr := range err.(validator.ValidationErrors) {
		field := validationErr.Field()
		tag := validationErr.Tag()
		param := validationErr.Param()

		template, found := errorMessages[tag]
		if !found {
			template = "is not valid."
		}

		var msg string
		if strings.Contains(template, "%s") {
			msg = fmt.Sprintf("Field '%s' "+template, field, param)
		} else {
			msg = fmt.Sprintf("Field '%s' %s", field, template)
		}

		messages = append(messages, msg)
	}

	return strings.Join(messages, " ")
}
