package request

import "time"

// CreateTeacherRequest describes payload for teacher provisioning.
type CreateTeacherRequest struct {
	INS        string  `json:"ins" binding:"required"`
	Password   string  `json:"password" binding:"required,min=8"`
	Email      *string `json:"email,omitempty"`
	FirstName  string  `json:"firstName" binding:"required"`
	LastName   string  `json:"lastName" binding:"required"`
	MiddleName *string `json:"middleName,omitempty"`
	Title      *string `json:"title,omitempty"`
	Bio        *string `json:"bio,omitempty"`
}

// CreateStudentRequest provisions a student account.
type CreateStudentRequest struct {
	INS        string  `json:"ins" binding:"required"`
	Index      string  `json:"index" binding:"required"`
	Password   string  `json:"password" binding:"required,min=8"`
	Email      *string `json:"email,omitempty"`
	FirstName  string  `json:"firstName" binding:"required"`
	LastName   string  `json:"lastName" binding:"required"`
	MiddleName *string `json:"middleName,omitempty"`
	GroupID    *string `json:"groupId,omitempty"`
}

// CreateGroupRequest creates a group.
type CreateGroupRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
}

// CreateSubjectRequest defines a subject.
type CreateSubjectRequest struct {
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
}

// AssignTeacherRequest assigns a teacher to a subject.
type AssignTeacherRequest struct {
	TeacherID string `json:"teacherId" binding:"required"`
}

// AttachGroupRequest links group to subject.
type AttachGroupRequest struct {
	GroupID string `json:"groupId" binding:"required"`
}

// AssignStudentToGroupRequest moves a student to a group.
type AssignStudentToGroupRequest struct {
	StudentID string `json:"studentId" binding:"required"`
}

// ScheduleSessionRequest defines lesson scheduling.
type ScheduleSessionRequest struct {
	SubjectID string     `json:"subjectId" binding:"required"`
	GroupID   string     `json:"groupId" binding:"required"`
	TeacherID string     `json:"teacherId" binding:"required"`
	StartsAt  time.Time  `json:"startsAt" binding:"required"`
	EndsAt    *time.Time `json:"endsAt,omitempty"`
	Topic     *string    `json:"topic,omitempty"`
}
