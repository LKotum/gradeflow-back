package response

// JournalAttendanceEntry represents a student's attendance for a lesson
type JournalAttendanceEntry struct {
	LessonID string `json:"lessonId"`
	Status   string `json:"status"` // present|absent|late
}

// JournalGradeEntry represents a student's grade for an assessment
type JournalGradeEntry struct {
	AssessmentID string   `json:"assessmentId"`
	Type         string   `json:"type"`
	Scale        string   `json:"scale"`
	ValueNum     *float64 `json:"valueNum,omitempty"`
	ValuePass    *bool    `json:"valuePass,omitempty"`
}

// StudentJournal groups attendance and grades, optionally filtered by course/date
type StudentJournal struct {
	StudentID  string                   `json:"studentId"`
	CourseID   *string                  `json:"courseId,omitempty"`
	Attendance []JournalAttendanceEntry `json:"attendance"`
	Grades     []JournalGradeEntry      `json:"grades"`
}
