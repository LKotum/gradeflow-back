package request

import "time"

// CreateTeacherRequest describes payload for teacher provisioning.
type CreateTeacherRequest struct {
	Password   string  `json:"password" binding:"required,min=8"`
	Email      *string `json:"email,omitempty"`
	FirstName  string  `json:"firstName" binding:"required"`
	LastName   string  `json:"lastName" binding:"required"`
	MiddleName *string `json:"middleName,omitempty"`
	Title      *string `json:"title,omitempty"`
	Bio        *string `json:"bio,omitempty"`
}

// UpdateTeacherRequest updates teacher profile fields.
type UpdateTeacherRequest struct {
	Email      *string `json:"email,omitempty"`
	FirstName  *string `json:"firstName,omitempty"`
	LastName   *string `json:"lastName,omitempty"`
	MiddleName *string `json:"middleName,omitempty"`
	Title      *string `json:"title,omitempty"`
	Bio        *string `json:"bio,omitempty"`
}

// CreateStudentRequest provisions a student account.
type CreateStudentRequest struct {
	Password   string  `json:"password" binding:"required,min=8"`
	Email      *string `json:"email,omitempty"`
	FirstName  string  `json:"firstName" binding:"required"`
	LastName   string  `json:"lastName" binding:"required"`
	MiddleName *string `json:"middleName,omitempty"`
	GroupID    *string `json:"groupId,omitempty"`
}

// UpdateStudentRequest updates student profile information.
type UpdateStudentRequest struct {
	Email      *string `json:"email,omitempty"`
	FirstName  *string `json:"firstName,omitempty"`
	LastName   *string `json:"lastName,omitempty"`
	MiddleName *string `json:"middleName,omitempty"`
	GroupID    *string `json:"groupId,omitempty"`
}

// ScheduleQuery describes filters for timetable requests.
type ScheduleQuery struct {
	SubjectID *string    `form:"subjectId"`
	GroupID   *string    `form:"groupId"`
	TeacherID *string    `form:"teacherId"`
	From      *time.Time `form:"from" time_format:"2006-01-02"`
	To        *time.Time `form:"to" time_format:"2006-01-02"`
}

// CreateGroupRequest creates a group.
type CreateGroupRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
}

// UpdateGroupRequest updates group metadata.
type UpdateGroupRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CreateSubjectRequest defines a subject.
type CreateSubjectRequest struct {
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
}

// UpdateSubjectRequest updates subject metadata.
type UpdateSubjectRequest struct {
	Code        *string `json:"code,omitempty"`
	Name        *string `json:"name,omitempty"`
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
	StudentIDs []string `json:"studentIds" binding:"required,min=1,dive,required"`
}

// ScheduleSessionRequest defines lesson scheduling.
type ScheduleSessionRequest struct {
	SubjectID string   `json:"subjectId" binding:"required"`
	GroupIDs  []string `json:"groupIds" binding:"required,min=1,dive,required"`
	TeacherID string   `json:"teacherId" binding:"required"`
	Date      string   `json:"date" binding:"required,datetime=2006-01-02"`
	Slot      int      `json:"slot" binding:"required,min=1,max=6"`
	Topic     *string  `json:"topic,omitempty"`
}

// ScheduleSessionUpdateRequest updates a scheduled lesson.
type ScheduleSessionUpdateRequest struct {
	GroupIDs  *[]string `json:"groupIds,omitempty"`
	TeacherID *string   `json:"teacherId,omitempty"`
	Date      *string   `json:"date,omitempty" binding:"omitempty,datetime=2006-01-02"`
	Slot      *int      `json:"slot,omitempty" binding:"omitempty,min=1,max=6"`
	Topic     *string   `json:"topic,omitempty"`
}

// UpdateGradeRequest allows dean staff to update a grade.
type UpdateGradeRequest struct {
	Value float32 `json:"value" binding:"required,gte=2,lte=5"`
	Notes *string `json:"notes,omitempty"`
}
