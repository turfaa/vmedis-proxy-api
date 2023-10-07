package schema

// User represents a user.
type User struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// LoginRequest represents the login request.
type LoginRequest struct {
	Email string `json:"email"`
}
