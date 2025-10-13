package models

// Запись на курс
type Enrollment struct {
	Base
	CourseID  string `gorm:"type:uuid;not null;index;uniqueIndex:uniq_enrollment_course_student,priority:1"`
	StudentID string `gorm:"type:uuid;not null;index;uniqueIndex:uniq_enrollment_course_student,priority:2"`
}
