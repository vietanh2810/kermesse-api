package request

import (
	"errors"
	"unicode"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

const (
	passwordMinLength = 8
)

var (
	errInvalidPassword         = errors.New("the password must be at least 8 characters and contain 1 letter, 1 number and 1 symbol")
	errConfirmPasswordMismatch = errors.New("confirm password doesn't match the password")
)

//type SignupRequest struct {
//	Email           string `json:"email" validate:"required"`
//	Password        string `json:"password" validate:"required"`
//	ConfirmPassword string `json:"confirm_password" validate:"required"`
//	Role            string `json:"role" binding:"required,oneof=student parent"`
//	StudentEmail    string `json:"student_email,omitempty"`
//	Name            string `json:"name,omitempty"`
//}

type SignupRequest struct {
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	Password        string   `json:"password"`
	ConfirmPassword string   `json:"confirm_password"`
	Role            string   `json:"role"`
	StudentEmails   []string `json:"student_emails,omitempty"`
}

func isPasswordValid(password string) bool {
	hasLetter := false
	hasNumber := false
	hasSymbol := false

	for _, char := range password {
		switch {
		case unicode.IsLetter(char):
			hasLetter = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}

		if hasLetter && hasNumber && hasSymbol {
			return true
		}
	}

	return false
}

func (req *SignupRequest) Validate() error {
	// Common validation for all roles
	err := validation.ValidateStruct(req,
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required, validation.Length(passwordMinLength, 0)),
		validation.Field(&req.ConfirmPassword, validation.Required),
		validation.Field(&req.Name, validation.Required),
		validation.Field(&req.Role, validation.Required, validation.In("student", "parent", "stand_holder", "organizer")),
	)
	if err != nil {
		return err
	}

	if !isPasswordValid(req.Password) {
		return errInvalidPassword
	}

	if req.Password != req.ConfirmPassword {
		return errConfirmPasswordMismatch
	}

	// Role-specific validation
	switch req.Role {
	case "parent":
		return validation.ValidateStruct(req,
			validation.Field(&req.StudentEmails, validation.Required, validation.Length(1, 0), validation.Each(is.Email)),
		)
	}

	return nil
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (req *LoginRequest) Validate() error {
	return validation.ValidateStruct(
		req,
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required),
	)
}
