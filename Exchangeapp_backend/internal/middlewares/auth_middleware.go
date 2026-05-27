package middlewares

import (
	"exchangeapp/pkg/jwtauth"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleWare(jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"Error": "Missing Authorization Header"})
			ctx.Abort()
			return
		}

		claims, err := jwtauth.ParseJWT(token, jwtSecret)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"Error": "Invalid token"})
			ctx.Abort()
			return
		}

		ctx.Set("userID", claims.UserID)
		ctx.Set("username", claims.Username)
		ctx.Next()
	}
}
