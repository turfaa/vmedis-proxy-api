package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

const (
	emailHeader = "X-Email"
	userCtxKey  = "apotek-api-user"
)

var (
	guestUser = User{
		Email: "guest@auliafarma.com",
		Role:  RoleGuest,
	}
)

func GinMiddleware(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.GetHeader(emailHeader)
		if email == "" {
			SetGinContext(c, guestUser)
			c.Next()
			return
		}

		user, err := service.GetOrCreateUser(c, email)
		if err != nil {
			c.JSON(500, gin.H{
				"error": fmt.Sprintf("failed to get or create user: %s", err),
			})
			c.Abort()
			return
		}

		SetGinContext(c, user)
		c.Next()
	}
}

func FromGinContext(ctx *gin.Context) User {
	userAny, ok := ctx.Get(userCtxKey)
	if !ok {
		return guestUser
	}

	user, ok := userAny.(User)
	if !ok {
		return guestUser
	}

	return user
}

func SetGinContext(ctx *gin.Context, user User) {
	ctx.Set(userCtxKey, user)
}
