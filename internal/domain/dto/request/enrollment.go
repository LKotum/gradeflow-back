package request

type CreateEnrollment struct {
	CourseID  string `json:"courseId" binding:"required"`
	StudentID string `json:"studentId" binding:"required"`
}

type UpdateEnrollment struct {
	CourseID  string `json:"courseId"`
	StudentID string `json:"studentId"`
}

type ListEnrollmentQuery struct {
	PaginationQuery
	CourseID  string `form:"courseId"`
	StudentID string `form:"studentId"`
}
