package contextkeys

// ContextKey is a custom type for context keys to prevent collisions.
type ContextKey string

const (
	// UserID is the context key used to store and retrieve the user ID.
	UserID ContextKey = "userID"
)
