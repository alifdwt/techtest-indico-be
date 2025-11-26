package dto

import "github.com/go-playground/validator/v10"

func ValidateStruct(obj any) error {
	validate := validator.New()
	return validate.Struct(obj)
}
