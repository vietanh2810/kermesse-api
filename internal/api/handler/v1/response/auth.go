package response

import "github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"

type LoginResponse struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}
