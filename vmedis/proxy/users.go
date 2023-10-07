package proxy

import (
	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy/schema"
)

// Login handles the login request.
// Currently it's a dummy login method. It doesn't check the password.
func (s *ApiServer) HandleLogin(c *gin.Context) {
	var req schema.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "failed to parse request: " + err.Error(),
		})
		return
	}

	var user models.User
	s.DB.Where("email = ?", req.Email).First(&user)

	role := user.Role
	if role == "" {
		role = "guest"
	}

	c.JSON(200, schema.User{Email: req.Email, Role: role})
}
