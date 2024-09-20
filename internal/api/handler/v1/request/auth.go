package request

import (
	"errors"

	regexp "github.com/dlclark/regexp2"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

const (
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

var (
	errInvalidPassword         = errors.New("the password must be at least 8 characters and contain 1 letter, 1 number and 1 symbol")
	errConfirmPasswordMismatch = errors.New("confirm password doesn't match the password")
)

type SignupRequest struct {
	Email           string `json:"email" validate:"required"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

func (req *SignupRequest) Validate() error {
	err := validation.ValidateStruct(
		req,
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required),
		validation.Field(&req.ConfirmPassword, validation.Required),
	)
	if err != nil {
		return err
	}

	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	ok, err := passwordExp.MatchString(req.Password)
	if err != nil {
		return err
	}
	if !ok {
		return errInvalidPassword
	}

	if req.Password != req.ConfirmPassword {
		return errConfirmPasswordMismatch
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
