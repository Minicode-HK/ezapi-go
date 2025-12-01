package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT authentication provider
type JWTProvider struct {
	SecretKey string
}

func (p *JWTProvider) Authenticate(c *gin.Context) (*User, error) {
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
        return nil, errors.New("missing authorization header")
    }

    tokenString := strings.TrimPrefix(authHeader, "Bearer ")

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(p.SecretKey), nil
    })

    if err != nil || !token.Valid {
        return nil, errors.New("invalid token")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("invalid claims")
    }

    return &User{
        ID:   claims["sub"].(string),
        Role: claims["role"].(string),
    }, nil
}

func (p *JWTProvider) GenerateToken(userID, role string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":  userID,
        "role": role,
        "exp":  time.Now().Add(time.Hour * 24).Unix(),
    })
    return token.SignedString([]byte(p.SecretKey))
}

func NewJWTProvider(secretKey string) *JWTProvider {
	return &JWTProvider{SecretKey: secretKey}
}