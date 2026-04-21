package api

import (
	"fmt"
	"regexp"
	"strings"
)

type Validator struct {
	errors map[string]string
}

func NewValidator() *Validator {
	return &Validator{
		errors: make(map[string]string),
	}
}

func (v *Validator) Required(field string, value interface{}) *Validator {
	if value == nil {
		v.errors[field] = "Field is required"
		return v
	}

	switch val := value.(type) {
	case string:
		if strings.TrimSpace(val) == "" {
			v.errors[field] = "Field cannot be empty"
		}
	case []interface{}:
		if len(val) == 0 {
			v.errors[field] = "Field cannot be empty"
		}
	}
	return v
}

func (v *Validator) MinLength(field string, value string, min int) *Validator {
	if len(value) < min {
		v.errors[field] = fmt.Sprintf("Minimum length is %d characters", min)
	}
	return v
}

func (v *Validator) MaxLength(field string, value string, max int) *Validator {
	if len(value) > max {
		v.errors[field] = fmt.Sprintf("Maximum length is %d characters", max)
	}
	return v
}

func (v *Validator) Email(field string, value string) *Validator {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(strings.ToLower(value)) {
		v.errors[field] = "Invalid email format"
	}
	return v
}

func (v *Validator) Pattern(field string, value string, pattern string, message string) *Validator {
	matched, _ := regexp.MatchString(pattern, value)
	if !matched {
		v.errors[field] = message
	}
	return v
}

func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

func (v *Validator) Errors() map[string]string {
	return v.errors
}
