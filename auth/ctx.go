package auth

import "context"

// FromContext will also extract user from gin context
// that is set by GinMiddleware using SetGinContext.
func FromContext(ctx context.Context) User {
	user, ok := ctx.Value(userCtxKey).(User)
	if !ok {
		return guestUser
	}

	return user
}
