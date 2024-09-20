package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api/handler/v1/response"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/pkg/jwthelper"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/service"
)

type UserService interface {
	GetUser(ctx context.Context, id uint) (domain.User, error)
}

type UserHandler struct {
	svc UserService
}

func NewUserHandler(svc UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

// HandleGetUser godoc
// @Summary      Get a user
// @Tags         users
// @Produce      json
// @Param        userID   path       int  true "user ID"
// @Success      200      {object}   domain.User
// @Failure      401      {object}   response.Err
// @Failure      500      {object}   response.Err
// @Router       /users/{userID} [get]
func (h *UserHandler) HandleGetUser(ctx *gin.Context) {
	rawUserID := ctx.Param("userID")
	userID, err := strconv.Atoi(rawUserID)
	if err != nil {
		response.RenderErr(ctx, response.ErrInvalidInput("userID", rawUserID))

		return
	}

	if userID <= 0 {
		response.RenderErr(ctx, response.ErrNotFound("user", "ID", userID))

		return
	}

	claims, err := jwthelper.RetrieveClaimsFromContext(ctx)
	if err != nil {
		response.RenderErr(ctx, response.ErrInternalServerError(err))

		return
	}

	if uint(userID) != claims.UserID {
		response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("can't view user %v by user %v", userID, claims.UserID)))

		return
	}

	user, err := h.svc.GetUser(ctx.Request.Context(), uint(userID))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.RenderErr(ctx, response.ErrNotFound("user", "ID", userID))

			return
		}

		err = fmt.Errorf("v1.HandleGetUser -> h.svc.GetUser -> %w", err)
		response.RenderErr(ctx, response.ErrInternalServerError(err))

		return
	}

	ctx.JSON(http.StatusOK, user)
}
