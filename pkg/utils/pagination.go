package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ApplyPagination applies limit/offset from query parameters to the GORM query.
// Query params: limit, offset. Values <= 0 are ignored. Limit is clamped by maxLimit if > 0.
// defaultLimit is used when limit is not provided or invalid. Set to 0 to disable default limit.
func ApplyPagination(q *gorm.DB, c *gin.Context, maxLimit, defaultLimit int) *gorm.DB {
	limit := defaultLimit
	if ls := c.Query("limit"); ls != "" {
		if v, err := strconv.Atoi(ls); err == nil && v > 0 {
			if maxLimit > 0 && v > maxLimit {
				limit = maxLimit
			} else {
				limit = v
			}
		}
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if os := c.Query("offset"); os != "" {
		if v, err := strconv.Atoi(os); err == nil && v >= 0 {
			q = q.Offset(v)
		}
	}
	return q
}
