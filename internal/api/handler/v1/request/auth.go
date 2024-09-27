package request

import (
	"errors"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

const (
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d).{8,}$`
)

var (
	errInvalidPassword         = errors.New("the password must be at least 8 characters and contain 1 letter, 1 number and 1 symbol")
	errConfirmPasswordMismatch = errors.New("confirm password doesn't match the password")
	errMissingStudentEmail     = errors.New("student email is required for parent signup")
)

type SignupRequest struct {
	Email           string `json:"email" validate:"required"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
	Role            string `json:"role" binding:"required,oneof=student parent"`
	StudentEmail    string `json:"student_email,omitempty"`
	Name            string `json:"name,omitempty"`
}

type BaseSignupRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
	Name            string `json:"name" validate:"required"`
	Role            string `json:"role" validate:"required,oneof=student parent stand_holder organizer"`
}

type StudentSignupRequest struct {
	BaseSignupRequest
}

type ParentSignupRequest struct {
	BaseSignupRequest
	StudentEmail string `json:"student_email" validate:"required,email"`
}

type StandHolderSignupRequest struct {
	BaseSignupRequest
	StandName        string `json:"stand_name" validate:"required"`
	StandType        string `json:"stand_type" validate:"required,oneof=Food Drink Activity"`
	StandDescription string `json:"stand_description"`
	StandKermesse    uint   `json:"stand_kermesse" validate:"required"`
}

type OrganizerSignupRequest struct {
	BaseSignupRequest
	PhoneNumber string `json:"phone_number" validate:"required"`
}

func (req *BaseSignupRequest) Validate() error {
	err := validation.ValidateStruct(
		req,
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required),
		validation.Field(&req.ConfirmPassword, validation.Required),
		validation.Field(&req.Name, validation.Required),
		validation.Field(&req.Role, validation.Required, validation.In("student", "parent", "stand_holder", "organizer")),
	)
	if err != nil {
		return err
	}

	passwordExp := regexp.MustCompile(passwordRegexPattern)
	if !passwordExp.MatchString(req.Password) {
		return errInvalidPassword
	}

	if req.Password != req.ConfirmPassword {
		return errConfirmPasswordMismatch
	}

	return nil
}

func (req *ParentSignupRequest) Validate() error {
	if err := req.BaseSignupRequest.Validate(); err != nil {
		return err
	}
	return validation.ValidateStruct(req,
		validation.Field(&req.StudentEmail, validation.Required, is.Email),
	)
}

func (req *StandHolderSignupRequest) Validate() error {
	if err := req.BaseSignupRequest.Validate(); err != nil {
		return err
	}
	return validation.ValidateStruct(req,
		validation.Field(&req.StandName, validation.Required, validation.Length(2, 50)),
		validation.Field(&req.StandType, validation.Required, validation.In("food", "drink", "activity")),
		validation.Field(&req.StandDescription, validation.Length(0, 200)),
		validation.Field(&req.StandKermesse, validation.Required),
	)
}

func (req *OrganizerSignupRequest) Validate() error {
	if err := req.BaseSignupRequest.Validate(); err != nil {
		return err
	}
	return validation.ValidateStruct(req,
		validation.Field(&req.PhoneNumber, validation.Required, validation.Match(regexp.MustCompile(`^\+?[1-9]\d{1,14}$`))),
	)
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
