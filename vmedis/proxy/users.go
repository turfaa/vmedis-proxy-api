package proxy

import (
	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy/schema"
)

// Login handles the login request.
// Currently it's a dummy login method. It doesn't do any authentication.
func (s *ApiServer) HandleLogin(c *gin.Context) {
	var req schema.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "failed to parse request: " + err.Error(),
		})
		return
	}

	// Get the user, or create a guest user.
	var user models.User
	s.DB.Where(models.User{Email: req.Email}).Attrs(models.User{Role: "guest"}).FirstOrCreate(&user)

	c.JSON(200, schema.FromModelsUser(user))
}
