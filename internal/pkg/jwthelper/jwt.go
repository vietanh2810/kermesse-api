package jwthelper

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	expirationTime    = time.Hour
	signingMethod     = jwt.SigningMethodHS512
	errClaimsNotFound = errors.New("claims not found in the context")
	errClaimsNotValid = errors.New("claims not valid")
)

type Claims struct {
	jwt.RegisteredClaims

	UserID    uint
	UserAgent string
}

func GenerateToken(signingKey []byte, userID uint, userAgent string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expirationTime)),
		},
		UserID:    userID,
		UserAgent: userAgent,
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	tokenStr, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func ParseWithClaims(signingKey, tokenString string, claims *Claims) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(signingKey), nil
	})
}

func RetrieveClaimsFromContext(ctx *gin.Context) (*Claims, error) {
	val, exists := ctx.Get("claims")
	if !exists {
		return nil, errClaimsNotFound
	}

	claims, ok := val.(*Claims)
	if !ok {
		return nil, errClaimsNotValid
	}

	return claims, nil
}
