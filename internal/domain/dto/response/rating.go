package response

// GroupRatingEntry represents ranking metrics for a group.
type GroupRatingEntry struct {
	GroupID        string   `json:"groupId"`
	GroupName      *string  `json:"groupName,omitempty"`
	AverageGrade   float64  `json:"averageGrade"`
	PassRate       float64  `json:"passRate"`
	AttendanceRate float64  `json:"attendanceRate"`
	StudentCount   int      `json:"studentCount"`
	Rank           int      `json:"rank"`
}

// StudentRatingEntry represents ranking metrics for a student.
type StudentRatingEntry struct {
	StudentID      string   `json:"studentId"`
	FullName       string   `json:"fullName"`
	GroupID        *string  `json:"groupId,omitempty"`
	AverageGrade   float64  `json:"averageGrade"`
	PassRate       float64  `json:"passRate"`
	AttendanceRate float64  `json:"attendanceRate"`
	Rank           int      `json:"rank"`
}
