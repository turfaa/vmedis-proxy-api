package token

type InsertTokenRequest struct {
	Token string `json:"token"`
}

// RefreshStatus represents the state of the token refresh.
type RefreshStatus string

const (
	// RefreshStatusIdle means no token refresh is in progress.
	RefreshStatusIdle RefreshStatus = "IDLE"
	// RefreshStatusRefreshing means tokens are currently being refreshed.
	RefreshStatusRefreshing RefreshStatus = "REFRESHING"
)

// RefreshStatusResponse is the response schema for the token refresh status API.
type RefreshStatusResponse struct {
	Status RefreshStatus `json:"status"`
}
