package auth

type User struct {
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleStaff    Role = "staff"
	RoleReseller Role = "reseller"
	RoleGuest    Role = "guest"
)

type LoginRequest struct {
	Email string `json:"email"`
}
