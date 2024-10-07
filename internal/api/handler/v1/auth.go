package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api/handler/v1/request"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api/handler/v1/response"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/config"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/pkg/jwthelper"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/service"
)

type AuthService interface {
	Signup(ctx context.Context, user domain.User) (domain.User, error)
	Login(ctx context.Context, email, password string) (domain.User, error)
	SignupStudent(ctx context.Context, student domain.Student) (domain.User, error)
	SignupParent(ctx context.Context, parent domain.Parent, studentEmails []string) (domain.User, error)
	SignupStandHolder(ctx context.Context, standHolder domain.StandHolder) (domain.User, error)
	SignupOrganizer(ctx context.Context, organizer domain.Organizer) (domain.User, error)
}

type AuthHandler struct {
	conf *config.APIConfig
	svc  AuthService
}

func NewAuthHandler(conf *config.APIConfig, svc AuthService) *AuthHandler {
	return &AuthHandler{
		conf: conf,
		svc:  svc,
	}
}

// HandleSignup godoc
// @Summary      Signup a new user
// @Tags         auth
// @Produce      json
// @Param        request   body      request.SignupRequest true "request body"
// @Success      201      {object}   domain.User
// @Failure      400      {object}   response.Err
// @Failure      500      {object}   response.Err
// @Router       /auth/signup [post]
func (h *AuthHandler) HandleSignup(ctx *gin.Context) {
	var req request.SignupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(err))
		return
	}

	if err := req.Validate(); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(err))
		return
	}

	var user domain.User
	var err error

	switch req.Role {
	case "student":
		user, err = h.svc.SignupStudent(ctx.Request.Context(), domain.Student{
			User: domain.User{
				Email:    req.Email,
				Password: req.Password,
				Name:     req.Name,
				Role:     "student",
			},
		})

	case "parent":
		user, err = h.svc.SignupParent(ctx.Request.Context(), domain.Parent{
			User: domain.User{
				Email:    req.Email,
				Password: req.Password,
				Name:     req.Name,
				Role:     "parent",
			},
		}, req.StudentEmails)

	case "stand_holder":
		user, err = h.svc.SignupStandHolder(ctx.Request.Context(), domain.StandHolder{
			User: domain.User{
				Email:    req.Email,
				Password: req.Password,
				Name:     req.Name,
				Role:     "stand_holder",
			},
		})

	case "organizer":
		user, err = h.svc.SignupOrganizer(ctx.Request.Context(), domain.Organizer{
			User: domain.User{
				Email:    req.Email,
				Password: req.Password,
				Name:     req.Name,
				Role:     "organizer",
			},
		})

	default:
		response.RenderErr(ctx, response.ErrBadRequest(errors.New("invalid role")))
		return
	}

	if err != nil {
		if errors.Is(err, service.ErrUserEmailExists) {
			response.RenderErr(ctx, response.ErrBadRequest(service.ErrUserEmailExists))
			return
		}
		if errors.Is(err, service.ErrStudentNotFound) {
			response.RenderErr(ctx, response.ErrBadRequest(service.ErrStudentNotFound))
			return
		}
		err = fmt.Errorf("v1.HandleSignup -> h.svc.Signup -> %w", err)
		response.RenderErr(ctx, response.ErrInternalServerError(err))
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

//func (h *AuthHandler) HandleSignup(ctx *gin.Context) {
//	var baseReq request.BaseSignupRequest
//	if err := ctx.ShouldBindJSON(&baseReq); err != nil {
//		response.RenderErr(ctx, response.ErrBadRequest(err))
//		return
//	}
//
//	if err := baseReq.Validate(); err != nil {
//		response.RenderErr(ctx, response.ErrBadRequest(err))
//		return
//	}
//
//	var user domain.User
//	var err error
//
//	switch baseReq.Role {
//	case "student":
//		var req request.StudentSignupRequest
//		if err := ctx.ShouldBindJSON(&req); err != nil {
//			response.RenderErr(ctx, response.ErrBadRequest(err))
//			return
//		}
//		if err := req.Validate(); err != nil {
//			response.RenderErr(ctx, response.ErrBadRequest(err))
//			return
//		}
//		user, err = h.svc.SignupStudent(ctx.Request.Context(), domain.Student{
//			User: domain.User{
//				Email:    req.Email,
//				Password: req.Password,
//				Name:     req.Name,
//				Role:     "student",
//			},
//		})
//
//	case "parent":
//		var req request.ParentSignupRequest
//		if err := ctx.ShouldBindJSON(&req); err != nil {
//			response.RenderErr(ctx, response.ErrBadRequest(err))
//			return
//		}
//		if err := req.Validate(); err != nil {
//			response.RenderErr(ctx, response.ErrBadRequest(err))
//			return
//		}
//		user, err = h.svc.SignupParent(ctx.Request.Context(), domain.Parent{
//			User: domain.User{
//				Email:    req.Email,
//				Password: req.Password,
//				Name:     req.Name,
//				Role:     "parent",
//			},
//		}, req.StudentEmail)
//
//	case "stand_holder":
//		var req request.StandHolderSignupRequest
//		if err := ctx.ShouldBindJSON(&req); err != nil {
//			response.RenderErr(ctx, response.ErrBadRequest(err))
//			return
//		}
//		if err := req.Validate(); err != nil {
//			response.RenderErr(ctx, response.ErrBadRequest(err))
//			return
//		}
//		user, err = h.svc.SignupStandHolder(ctx.Request.Context(), domain.StandHolder{
//			User: domain.User{
//				Email:    req.Email,
//				Password: req.Password,
//				Name:     req.Name,
//				Role:     "stand_holder",
//			},
//			Stand: domain.Stand{
//				Name:        req.StandName,
//				Type:        req.StandType,
//				Description: req.StandDescription,
//				KermesseID:  req.StandKermesse,
//			},
//		})
//
//	case "organizer":
//		var req request.OrganizerSignupRequest
//		if err := ctx.ShouldBindJSON(&req); err != nil {
//			response.RenderErr(ctx, response.ErrBadRequest(err))
//			return
//		}
//		if err := req.Validate(); err != nil {
//			response.RenderErr(ctx, response.ErrBadRequest(err))
//			return
//		}
//		user, err = h.svc.SignupOrganizer(ctx.Request.Context(), domain.Organizer{
//			User: domain.User{
//				Email:    req.Email,
//				Password: req.Password,
//				Name:     req.Name,
//				Role:     "organizer",
//			},
//		})
//
//	default:
//		response.RenderErr(ctx, response.ErrBadRequest(errors.New("invalid role")))
//		return
//	}
//
//	if err != nil {
//		if errors.Is(err, service.ErrUserEmailExists) {
//			response.RenderErr(ctx, response.ErrBadRequest(service.ErrUserEmailExists))
//			return
//		}
//		if errors.Is(err, service.ErrStudentNotFound) {
//			response.RenderErr(ctx, response.ErrBadRequest(service.ErrStudentNotFound))
//			return
//		}
//		err = fmt.Errorf("v1.HandleSignup -> h.svc.Signup -> %w", err)
//		response.RenderErr(ctx, response.ErrInternalServerError(err))
//		return
//	}
//
//	ctx.JSON(http.StatusCreated, user)
//}

// HandleLogin godoc
// @Summary      Login a user
// @Tags         auth
// @Produce      json
// @Param        request   body      request.LoginRequest true "request body"
// @Success      200      {object}   domain.User
// @Failure      401      {object}   response.Err
// @Failure      500      {object}   response.Err
// @Router       /auth/login [post]
func (h *AuthHandler) HandleLogin(ctx *gin.Context) {
	req := request.LoginRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(err))

		return
	}

	if err := req.Validate(); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(err))

		return
	}

	user, err := h.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrWrongPassword) {
			response.RenderErr(ctx, response.ErrWrongCredentials(err))

			return
		}

		err = fmt.Errorf("v1.HandleSignup -> h.svc.Login -> %w", err)
		response.RenderErr(ctx, response.ErrInternalServerError(err))

		return
	}

	token, err := jwthelper.GenerateToken([]byte(h.conf.JWTSigningKey), user.ID, ctx.Request.UserAgent())
	if err != nil {
		err = fmt.Errorf("v1.HandleSignup -> middleware.GenerateToken() -> %w", err)
		response.RenderErr(ctx, response.ErrInternalServerError(err))

		return
	}

	ctx.JSON(http.StatusOK, response.LoginResponse{
		Token: token,
		User:  user,
	})
}
