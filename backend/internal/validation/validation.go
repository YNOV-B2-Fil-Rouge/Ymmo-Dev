// Package validation registers the app's custom validators on Gin's
// binding engine. Centralizing them here keeps validation rules reusable
// across every DTO (DRY).
package validation

import (
	"regexp"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// A pragmatic phone pattern: optional leading +, then 6 to 20 characters
// made of digits, spaces, dots, dashes or parentheses.
// Accepts "+33 6 12 34 56 78", "0612345678", "01.23.45.67.89" — rejects "TEST".
var phoneRegex = regexp.MustCompile(`^\+?[0-9][0-9 .\-()]{5,19}$`)

var once sync.Once

// Register installs the custom validators. Safe to call multiple times.
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
		return true // emptiness is handled by the "omitempty" tag
	}
	return phoneRegex.MatchString(value)
}
