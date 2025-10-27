package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Context keys for downstream handlers.
const (
	ContextUserIDKey = "userID"
	ContextRoleKey   = "userRole"
)

// JWTAuth validates bearer token and injects subject/role into context.
func JWTAuth(secret []byte) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		raw := strings.TrimPrefix(authHeader, "Bearer ")
		claims := jwtRegisteredClaims{}
		token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
		if err != nil || !token.Valid || claims.Type != "access" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		ctx.Set(ContextUserIDKey, claims.Subject)
		ctx.Set(ContextRoleKey, claims.Role)
		ctx.Next()
	}
}

// RequireRoles ensures that only specific roles can access the handler.
func RequireRoles(allowed ...string) gin.HandlerFunc {
	set := make(map[string]struct{}, len(allowed))
	for _, role := range allowed {
		set[role] = struct{}{}
	}
	return func(ctx *gin.Context) {
		role, ok := ctx.Get(ContextRoleKey)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "role missing"})
			return
		}
		if _, ok := set[role.(string)]; !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
			return
		}
		ctx.Next()
	}
}

type jwtRegisteredClaims struct {
	jwt.RegisteredClaims
	Type string `json:"type"`
	Role string `json:"role"`
}
