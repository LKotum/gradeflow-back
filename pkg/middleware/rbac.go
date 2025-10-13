package middleware

import (
	"net/http"

	resp "gradeflow/internal/domain/dto/response"
	m "gradeflow/internal/domain/models"

	"github.com/gin-gonic/gin"
)

const (
	RoleStudent = "student"
	RoleTeacher = "teacher"
	RoleDean    = "dean"
	RoleAdmin   = "admin"
)

// RequireRoles allows the request only if the current user's role is in allowed list.
func RequireRoles(allowed ...string) gin.HandlerFunc {
	set := map[string]struct{}{}
	for _, r := range allowed {
		set[r] = struct{}{}
	}
	return func(c *gin.Context) {
		v, ok := c.Get("user")
		if !ok || v == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, resp.Error{Error: "unauthorized"})
			return
		}
		u, ok := v.(*m.User)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, resp.Error{Error: "unauthorized"})
			return
		}
		if _, ok := set[u.Role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, resp.Error{Error: "forbidden"})
			return
		}
		c.Next()
	}
}

// Helpers to compose common policies
func OnlyAdmin() gin.HandlerFunc      { return RequireRoles(RoleAdmin) }
func AdminOrDean() gin.HandlerFunc    { return RequireRoles(RoleAdmin, RoleDean) }
func StaffOrTeacher() gin.HandlerFunc { return RequireRoles(RoleDean, RoleTeacher) }
func AnyAuthenticated() gin.HandlerFunc {
	return RequireRoles(RoleAdmin, RoleDean, RoleTeacher, RoleStudent)
}
