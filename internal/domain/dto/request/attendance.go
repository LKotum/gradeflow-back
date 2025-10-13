package request

type CreateAttendance struct {
	LessonID  string `json:"lessonId" binding:"required"`
	StudentID string `json:"studentId" binding:"required"`
	Status    string `json:"status" binding:"required"`
}

type UpdateAttendance struct {
	Status string `json:"status"`
}

type ListAttendanceQuery struct {
	PaginationQuery
	LessonID  string `form:"lessonId"`
	StudentID string `form:"studentId"`
}

type AttendanceItem struct {
	StudentID string `json:"studentId" binding:"required"`
	Status    string `json:"status" binding:"required"`
}
