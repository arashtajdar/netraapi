package contextkeys

import "context"

// contextKey is an unexported type to prevent collisions with context keys
// defined in other packages. Using a struct type rather than a string
// guarantees uniqueness across package boundaries.
type contextKey struct {
	name string
}

var (
	// UserIDKey is the context key for the authenticated user's ID.
	UserIDKey = &contextKey{"user_id"}

	// ProfileIDKey is the context key for the active profile ID.
	ProfileIDKey = &contextKey{"profile_id"}

	// UserLevelKey is the context key for the user's access level.
	UserLevelKey = &contextKey{"user_level"}

	// RequestIDKey is the context key for request tracking.
	RequestIDKey = &contextKey{"request_id"}
)

// WithUserID returns a new context with the user ID set.
func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithProfileID returns a new context with the profile ID set.
func WithProfileID(ctx context.Context, profileID int) context.Context {
	return context.WithValue(ctx, ProfileIDKey, profileID)
}

// WithRequestID returns a new context with the request ID set.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// UserIDFromContext safely extracts the user ID from the context.
// Returns (0, false) if the value is missing or not an int.
func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(UserIDKey).(int)
	return id, ok
}

// ProfileIDFromContext safely extracts the profile ID from the context.
// Returns (0, false) if the value is missing or not an int.
func ProfileIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(ProfileIDKey).(int)
	return id, ok
}

// RequestIDFromContext safely extracts the request ID from the context.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(RequestIDKey).(string)
	return id, ok
}

// WithUserLevel returns a new context with the user level set.
func WithUserLevel(ctx context.Context, level int) context.Context {
	return context.WithValue(ctx, UserLevelKey, level)
}

// UserLevelFromContext safely extracts the user level from the context.
// Returns 1 (default level) if not found.
func UserLevelFromContext(ctx context.Context) int {
	level, ok := ctx.Value(UserLevelKey).(int)
	if !ok {
		return 1 // Default level for unauthenticated users
	}
	return level
}
