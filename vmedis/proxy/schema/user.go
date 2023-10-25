package schema

import "github.com/turfaa/vmedis-proxy-api/vmedis/database/models"

// User represents a user.
type User struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// LoginRequest represents the login request.
type LoginRequest struct {
	Email string `json:"email"`
}

// FromModelsUser converts models.User to User.
func FromModelsUser(user models.User) User {
	return User{
		Email: user.Email,
		Role:  user.Role,
	}
}
