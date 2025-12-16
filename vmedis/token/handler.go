package token

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/turfaa/vmedis-proxy-api/cui"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/time2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetTokens(c *gin.Context) {
	tokens, err := h.service.GetTokens(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get tokens: %s", err),
		})
		return
	}

	c.JSON(200, h.transformTokensToTable(tokens))
}

func (h *Handler) transformTokensToTable(tokens []models.VmedisToken) cui.Table {
	header := []string{
		"Diinput",
		"Terakhir Diperbarui",
		"Token",
		"Status",
	}

	rows := make([]cui.Row, len(tokens))
	for i, token := range tokens {
		rows[i] = cui.Row{
			ID: strconv.FormatUint(uint64(token.ID), 10),
			Columns: []string{
				time2.FormatDateTime(token.CreatedAt),
				time2.FormatDateTime(token.UpdatedAt),
				token.Token,
				token.State.String(),
			},
		}
	}

	return cui.Table{
		Header: header,
		Rows:   rows,
	}
}

func (h *Handler) InsertToken(c *gin.Context) {
	var request InsertTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("failed to parse request: %s", err),
		})
		return
	}

	if err := h.service.InsertToken(c, request.Token); err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to insert token: %s", err),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Token inserted successfully",
	})
}

func (h *Handler) DeleteToken(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("failed to parse id: %s", err),
		})
		return
	}

	if err := h.service.DeleteToken(c, uint(id)); err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to delete token: %s", err),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Token deleted successfully",
	})
}
