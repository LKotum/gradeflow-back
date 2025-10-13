package response

// StudentSummaryReport aggregates grades and attendance for a student.
type StudentSummaryReport struct {
	StudentID        string  `json:"studentId"`
	AverageGrade     float64 `json:"averageGrade"`
	PassRate         float64 `json:"passRate"`
	AttendanceRate   float64 `json:"attendanceRate"`
	TotalGrades      int     `json:"totalGrades"`
	TotalAttendances int     `json:"totalAttendances"`
}

// CourseSummaryReport aggregates performance metrics for a course.
type CourseSummaryReport struct {
	CourseID         string  `json:"courseId"`
	AverageGrade     float64 `json:"averageGrade"`
	PassRate         float64 `json:"passRate"`
	AttendanceRate   float64 `json:"attendanceRate"`
	TotalGrades      int     `json:"totalGrades"`
	StudentCount     int     `json:"studentCount"`
	UniqueGroups     int     `json:"uniqueGroups"`
}

// SessionSummaryReport aggregates metrics for an academic session.
type SessionSummaryReport struct {
	SessionID        string  `json:"sessionId"`
	AverageGrade     float64 `json:"averageGrade"`
	PassRate         float64 `json:"passRate"`
	AttendanceRate   float64 `json:"attendanceRate"`
	TotalGrades      int     `json:"totalGrades"`
	StudentCount     int     `json:"studentCount"`
	CourseCount      int     `json:"courseCount"`
}
