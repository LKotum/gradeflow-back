package response

import "time"

// UserProfile represents basic identity information.
type UserProfile struct {
	ID         string  `json:"id"`
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	MiddleName *string `json:"middleName,omitempty"`
	Email      *string `json:"email,omitempty"`
	INS        *string `json:"ins,omitempty"`
	AvatarURL  *string `json:"avatarUrl,omitempty"`
	Role       string  `json:"role"`
}

// GroupSummary contains lightweight group info.
type GroupSummary struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// SubjectSummary holds subject details.
type SubjectSummary struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// TeacherSubjectSummary describes subject and attached groups.
type TeacherSubjectSummary struct {
	Subject SubjectSummary `json:"subject"`
	Groups  []GroupSummary `json:"groups"`
}

// TeacherDashboardResponse aggregates data for teacher cabinet.
type TeacherDashboardResponse struct {
	Profile  UserProfile              `json:"profile"`
	Subjects []TeacherSubjectSummary  `json:"subjects"`
}

// SessionSummary provides lesson metadata.
type SessionSummary struct {
	ID        string     `json:"id"`
	StartsAt  time.Time  `json:"startsAt"`
	EndsAt    *time.Time `json:"endsAt,omitempty"`
	Topic     *string    `json:"topic,omitempty"`
	SubjectID string     `json:"subjectId"`
	GroupID   string     `json:"groupId"`
}

// GradeDetail contains grade info per student/session.
type GradeDetail struct {
	GradeID   *string       `json:"gradeId,omitempty"`
	SessionID string        `json:"sessionId"`
	Student   UserProfile   `json:"student"`
	Value     *float32      `json:"value,omitempty"`
	Notes     *string       `json:"notes,omitempty"`
	AssessedAt *time.Time   `json:"assessedAt,omitempty"`
}

// GradeTableResponse is used for teacher gradebook view.
type GradeTableResponse struct {
	Subject SubjectSummary `json:"subject"`
	Group   GroupSummary   `json:"group"`
	Sessions []SessionSummary `json:"sessions"`
	Students []UserProfile    `json:"students"`
	Grades   []GradeDetail    `json:"grades"`
}

// StudentSubjectGrade groups student grades by subject.
type StudentSubjectGrade struct {
	Subject  SubjectSummary        `json:"subject"`
	Sessions []StudentSessionGrade `json:"sessions"`
	Average  *float32              `json:"average,omitempty"`
}

// StudentSessionGrade exposes grade for a specific session.
type StudentSessionGrade struct {
	Session SessionSummary `json:"session"`
	Grade   *float32       `json:"grade,omitempty"`
	GradeID *string        `json:"gradeId,omitempty"`
	Notes   *string        `json:"notes,omitempty"`
}

// StudentDashboardResponse contains student cabinet info.
type StudentDashboardResponse struct {
	Profile     UserProfile  `json:"profile"`
	Group       *GroupSummary `json:"group,omitempty"`
	AverageGPA  *float32     `json:"averageGpa,omitempty"`
}

// AverageMetricResponse represents aggregated score metrics.
type AverageMetricResponse struct {
	SubjectAverage   *float32 `json:"subjectAverage,omitempty"`
	GroupAverage     *float32 `json:"groupAverage,omitempty"`
	OverallAverage   *float32 `json:"overallAverage,omitempty"`
}

// GroupRankingItem shows ranking info for a group.
type GroupRankingItem struct {
	Group   GroupSummary `json:"group"`
	Average *float32     `json:"average,omitempty"`
}

// GroupRankingResponse lists ranking by groups.
type GroupRankingResponse struct {
	Items []GroupRankingItem `json:"items"`
}
