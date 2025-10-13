package response

type Attendance struct {
	ID        string  `json:"id"`
	LessonID  string  `json:"lessonId"`
	StudentID string  `json:"studentId"`
	Status    string  `json:"status"`
	MarkedBy  *string `json:"markedBy"`
	MarkedAt  string  `json:"markedAt"`
}
