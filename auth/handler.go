package auth

import "github.com/gin-gonic/gin"

type ApiHandler struct {
	service *Service
}

func (h *ApiHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "failed to parse request: " + err.Error(),
		})
		return
	}

	user, err := h.service.GetOrCreateUser(c, req.Email)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "failed to login: " + err.Error(),
		})
		return
	}

	c.JSON(200, user)
}

func NewApiHandler(service *Service) *ApiHandler {
	return &ApiHandler{service: service}
}
