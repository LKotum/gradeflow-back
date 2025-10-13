package request

import "time"

type CreateAcademicSession struct {
	Code     string    `json:"code" binding:"required"`
	Kind     string    `json:"kind" binding:"required"`
	StartsAt time.Time `json:"startsAt" binding:"required"`
	EndsAt   time.Time `json:"endsAt" binding:"required"`
}

type UpdateAcademicSession struct {
	Code     string     `json:"code"`
	Kind     string     `json:"kind"`
	StartsAt *time.Time `json:"startsAt"`
	EndsAt   *time.Time `json:"endsAt"`
}

type ListAcademicSessionQuery struct{ PaginationQuery }
