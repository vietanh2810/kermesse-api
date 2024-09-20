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
	req := request.SignupRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(err))

		return
	}

	if err := req.Validate(); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(err))

		return
	}

	user, err := h.svc.Signup(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrUserEmailExists) {
			response.RenderErr(ctx, response.ErrBadRequest(service.ErrUserEmailExists))

			return
		}

		err = fmt.Errorf("v1.HandleSignup -> h.svc.Signup -> %w", err)
		response.RenderErr(ctx, response.ErrInternalServerError(err))

		return
	}

	ctx.JSON(http.StatusCreated, user)
}

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
