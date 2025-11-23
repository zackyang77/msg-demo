package authctx

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const userIDKey contextKey = "msgdemo:userId"

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromCtx(ctx context.Context) (int64, bool) {
	val := ctx.Value(userIDKey)
	if val == nil {
		return 0, false
	}

	if id, ok := val.(int64); ok {
		return id, true
	}

	return 0, false
}

func UserIDFromClaims(claims jwt.MapClaims) (int64, bool) {
	value, ok := claims["userId"]
	if !ok {
		return 0, false
	}

	switch v := value.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int32:
		return int64(v), true
	default:
		return 0, false
	}
}
