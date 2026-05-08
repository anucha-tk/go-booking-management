package api

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// FormatValidationError formats validator.ValidationErrors into a map.
func FormatValidationError(err error) map[string]string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make(map[string]string)
		for _, fe := range ve {
			out[fe.Field()] = getErrorMsg(fe)
		}
		return out
	}
	return nil
}

func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Should be at least %s characters long", fe.Param())
	case "max":
		return fmt.Sprintf("Should be at most %s characters long", fe.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", fe.Param())
	}
	return fmt.Sprintf("Invalid value for %s", fe.Field())
}

// BindAndValidate binds the request to the struct and handles validation errors.
// It uses Gin's default validator which looks for 'binding' tags.
func BindAndValidate(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		verrs := FormatValidationError(err)
		if verrs != nil {
			ValidationError(c, verrs)
			return false
		}
		Error(c, 400, "INVALID_REQUEST", "Invalid request payload")
		return false
	}
	return true
}

// Validate validates a struct using 'validate' tags (Service Layer).
func Validate(s interface{}) error {
	return validate.Struct(s)
}

// GetValidator returns the shared validator instance.
func GetValidator() *validator.Validate {
	return validate
}

// RegisterTagNameFunc registers a function that gets the JSON tag name for validator errors.
func RegisterTagNameFunc(v *validator.Validate) {
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// InitValidator initializes the Gin validator and internal validator to use JSON tag names.
func InitValidator() {
	// Register for Gin's validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		RegisterTagNameFunc(v)
	}
	// Register for shared internal validator
	RegisterTagNameFunc(validate)
}
