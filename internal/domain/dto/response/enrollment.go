package response

type Enrollment struct {
	ID        string `json:"id"`
	CourseID  string `json:"courseId"`
	StudentID string `json:"studentId"`
}
