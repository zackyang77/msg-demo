package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/pineapple/msg-demo/backend/inbox/internal/pkg/authctx"
)

type AuthMiddleware struct {
	secret []byte
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{
		secret: []byte(secret),
	}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractToken(r)
		if tokenStr == "" {
			writeUnauthorized(r, w, "缺少身份凭证")
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return m.secret, nil
		})
		if err != nil || !token.Valid {
			writeUnauthorized(r, w, "身份凭证无效")
			return
		}

		userID, ok := authctx.UserIDFromClaims(claims)
		if !ok {
			writeUnauthorized(r, w, "身份信息缺失")
			return
		}

		ctx := authctx.WithUserID(r.Context(), userID)
		next(w, r.WithContext(ctx))
	}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func writeUnauthorized(r *http.Request, w http.ResponseWriter, message string) {
	httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, map[string]string{
		"message": message,
	})
}
