// Package validation registers the app's custom validators on Gin.
package validation

import (
	"regexp"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var phoneRegex = regexp.MustCompile(`^\+?[0-9][0-9 .\-()]{5,19}$`)

var once sync.Once

func Register() {
	once.Do(func() {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			_ = v.RegisterValidation("phone", validatePhone)
		}
	})
}

func validatePhone(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	return phoneRegex.MatchString(value)
}
