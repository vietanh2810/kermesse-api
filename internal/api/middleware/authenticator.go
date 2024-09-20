package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api/handler/v1/response"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/pkg/jwthelper"
)

type Authenticator struct {
	JWTSigningKey string
}

func NewAuthenticator(jwtSigningKey string) *Authenticator {
	return &Authenticator{
		JWTSigningKey: jwtSigningKey,
	}
}

func (a *Authenticator) VerifyJWT() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, err := a.extractClaims(ctx)
		if err != nil {
			response.RenderErr(ctx, response.ErrJWTUnverified(err))

			return
		}

		ctx.Set("claims", claims)
	}
}

func (a *Authenticator) extractClaims(ctx *gin.Context) (*jwthelper.Claims, error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return nil, errors.New("header Authorization not found")
	}

	// We expect the header like "Authorization: BEARER T",
	// only the portion after BEARER is needed.
	if len(authHeader) <= 7 || strings.ToUpper(authHeader[0:6]) != "BEARER" {
		return nil, errors.New("wrong JWT format")
	}

	bearer := authHeader[7:]
	claims := &jwthelper.Claims{} // must pass pointer into ParseWithClaims.
	token, err := jwthelper.ParseWithClaims(a.JWTSigningKey, bearer, claims)
	if err != nil {
		return nil, err
	}

	if token == nil || !token.Valid {
		return nil, errors.New("token is invalid or already expired")
	}

	if claims.UserID == 0 {
		return nil, errors.New("userID not found in the claims")
	}

	if claims.UserAgent != ctx.Request.UserAgent() {
		return nil, errors.New("user agent do not match")
	}

	return claims, nil
}
