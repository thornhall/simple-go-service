package auth

// import (
// 	"net/http"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang-jwt/jwt/v4"
// 	"github.com/thornhall/simple-go-service/internal/service"
// )

// func JWTAuth(userSvc *service.UserService, jwtSecret []byte) gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		auth := ctx.GetHeader("Authorization")
// 		parts := strings.SplitN(auth, " ", 2)
// 		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
// 			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
// 			return
// 		}

// 		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
// 			return jwtSecret, nil
// 		}, jwt.WithValidMethods([]string{"HS256"}))
// 		if err != nil || !token.Valid {
// 			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
// 			return
// 		}
// 		claims, ok := token.Claims.(jwt.MapClaims)
// 		if !ok {
// 			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
// 			return
// 		}

// 		sub, ok := claims["sub"].(string)
// 		if !ok {
// 			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token missing sub claim"})
// 			return
// 		}

// 		userResp, err := userSvc.Get(sub)
// 		if err != nil {
// 			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
// 			return
// 		}

// 		ctx.Set("userID", sub)
// 		ctx.Set("currentUser", userResp)

// 		ctx.Next()
// 	}
// }
