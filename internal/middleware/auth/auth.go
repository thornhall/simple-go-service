package auth

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func IssueJWT(userID int64, email string) (string, error) {
	strUserId := strconv.FormatInt(userID, 10)
	claims := jwt.MapClaims{
		"sub":   strUserId,
		"email": email,
		"exp":   time.Now().Add(15 * 24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("encountered an error when creating user")
	}
	return token.SignedString([]byte(jwtSecret))
}

func JWTAuth(jwtSecret []byte) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth := ctx.GetHeader("Authorization")
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			ctx.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "missing or invalid Authorization header"},
			)
			return
		}

		token, err := jwt.ParseWithClaims(
			parts[1],
			&jwt.RegisteredClaims{},
			func(t *jwt.Token) (any, error) {
				if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, jwt.ErrTokenSignatureInvalid
				}
				return jwtSecret, nil
			},
			jwt.WithValidMethods([]string{"HS256"}),
		)
		if err != nil || !token.Valid {
			log.Println(err.Error())
			ctx.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "missing or invalid Authorization header"},
			)
			return
		}
		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok {
			log.Println("claims is not of type *MyClaims")
			ctx.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "missing or invalid Authorization header"},
			)
			return
		}

		if claims.Subject == "" {
			log.Println("sub is empty")
			ctx.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "missing or invalid Authorization header"},
			)
			return
		}

		ctx.Set("userId", claims.Subject)
		ctx.Next()
	}
}
