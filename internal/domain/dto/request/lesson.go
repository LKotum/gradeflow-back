package request

import "time"

type CreateLesson struct {
	CourseID string    `json:"courseId" binding:"required"`
	StartsAt time.Time `json:"startsAt" binding:"required"`
	EndsAt   time.Time `json:"endsAt" binding:"required"`
	Room     string    `json:"room"`
	Kind     string    `json:"kind" binding:"required"`
}

type UpdateLesson struct {
	CourseID string     `json:"courseId"`
	StartsAt *time.Time `json:"startsAt"`
	EndsAt   *time.Time `json:"endsAt"`
	Room     string     `json:"room"`
	Kind     string     `json:"kind"`
}

type ListLessonQuery struct {
	PaginationQuery
	CourseID  string `form:"courseId"`
	SessionID string `form:"sessionId"`
	From      string `form:"from"` // RFC3339
	To        string `form:"to"`   // RFC3339
}

type AttendanceBulkItem struct {
	StudentID string `json:"studentId" binding:"required"`
	Status    string `json:"status" binding:"required"`
}
