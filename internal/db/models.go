// internal/db/models.go
package db

import (
	"time"

	"gorm.io/gorm"
)

// Базовая структура с UUID (pgcrypto)
type Base struct {
	ID        string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time      `gorm:"not null;default:now()"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Кафедра
type Department struct {
	Base
	Code string `gorm:"uniqueIndex;not null"`
	Name string `gorm:"not null;index"`
}

// Направление/программа
type Program struct {
	Base
	DepartmentID string     `gorm:"type:uuid;not null;index"`
	Department   Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:RESTRICT"`
	Code         string     `gorm:"uniqueIndex;not null"`
	Name         string     `gorm:"not null;index"`
}

// Предмет (дисциплина)
type Subject struct {
	Base
	DepartmentID string     `gorm:"type:uuid;not null;index"`
	Department   Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:RESTRICT"`
	Code         string     `gorm:"uniqueIndex;not null"`
	Title        string     `gorm:"not null;index"`
	Credits      int        `gorm:"not null;default:0"`
}

// Учебная сессия (период)
type AcademicSession struct {
	Base
	Code     string    `gorm:"uniqueIndex;not null"` // 2025-spring и т.п.
	Kind     string    `gorm:"not null;index"`       // teaching|exam
	StartsAt time.Time `gorm:"not null"`
	EndsAt   time.Time `gorm:"not null"`
}

// Курс (предложение предмета в конкретной сессии)
type Course struct {
	Base
	SubjectID         string          `gorm:"type:uuid;not null;index"`
	Subject           Subject         `gorm:"foreignKey:SubjectID;constraint:OnDelete:RESTRICT"`
	DepartmentID      string          `gorm:"type:uuid;not null;index"`
	Department        Department      `gorm:"foreignKey:DepartmentID;constraint:OnDelete:RESTRICT"`
	ProgramID         *string         `gorm:"type:uuid;index"`
	Program           *Program        `gorm:"foreignKey:ProgramID;constraint:OnDelete:SET NULL"`
	AcademicSessionID string          `gorm:"type:uuid;not null;index"`
	AcademicSession   AcademicSession `gorm:"foreignKey:AcademicSessionID;constraint:OnDelete:RESTRICT"`
	Title             string          `gorm:"not null;index"`
	TeacherID         string          `gorm:"type:text;index"` // внешняя ссылка
	Room              string          `gorm:"index"`
}

// Группа
type Group struct {
	Base
	ProgramID string  `gorm:"type:uuid;not null;index"`
	Program   Program `gorm:"foreignKey:ProgramID;constraint:OnDelete:RESTRICT"`
	Code      string  `gorm:"uniqueIndex;not null"`
	Name      string  `gorm:"not null;index"`
	Year      int     `gorm:"not null;index"`
}

// Студент (с индивидуальным номером)
type Student struct {
	Base
	IndividualNumber string  `gorm:"uniqueIndex;not null"` // ✅ вместо зачётной книжки
	FullName         string  `gorm:"not null;index"`
	GroupID          *string `gorm:"type:uuid;index"`
	Group            *Group  `gorm:"foreignKey:GroupID;constraint:OnDelete:SET NULL"`
}

// Запись на курс
type Enrollment struct {
	Base
	CourseID  string   `gorm:"type:uuid;not null;index"`
	StudentID string   `gorm:"type:uuid;not null;index"`
	_         struct{} `gorm:"uniqueIndex:uniq_enrollment_course_student,priority:1"`
	__        struct{} `gorm:"uniqueIndex:uniq_enrollment_course_student,priority:2"`
}

// Занятие (расписание)
type Lesson struct {
	Base
	CourseID string    `gorm:"type:uuid;not null;index"`
	StartsAt time.Time `gorm:"not null;index"`
	EndsAt   time.Time `gorm:"not null;index"`
	Room     string    `gorm:"index"`
	Kind     string    `gorm:"not null;index"` // lecture|seminar|lab
}

// Посещаемость
type Attendance struct {
	Base
	LessonID  string    `gorm:"type:uuid;not null;index"`
	StudentID string    `gorm:"type:uuid;not null;index"`
	Status    string    `gorm:"not null;index"` // present|absent|late
	MarkedBy  string    `gorm:"type:text;not null"`
	MarkedAt  time.Time `gorm:"not null;default:now()"`
	_         struct{}  `gorm:"uniqueIndex:uniq_attendance_lesson_student,priority:1"`
	__        struct{}  `gorm:"uniqueIndex:uniq_attendance_lesson_student,priority:2"`
}

// ОЦЕНОЧНОЕ МЕРОПРИЯТИЕ (экзамен/зачёт/дифф.зачёт) — «сессия (оценка)»
type Assessment struct {
	Base
	CourseID string    `gorm:"type:uuid;not null;index"`
	Type     string    `gorm:"not null;index"` // exam|zachet|diff_zachet
	DateAt   time.Time `gorm:"not null;index"`
	Room     string    `gorm:"index"`
	// шкала для оценок по данному мероприятию:
	// passfail (зачёт/незачёт) | five (2..5) | hundred (0..100)
	Scale  string `gorm:"not null;index"`
	MaxPts *float64
}

// Оценка по мероприятию (универсальная модель под зачёт/дифф.зачёт/экзамен)
type AssessmentGrade struct {
	Base
	AssessmentID string     `gorm:"type:uuid;not null;index"`
	Assessment   Assessment `gorm:"foreignKey:AssessmentID;constraint:OnDelete:CASCADE"`
	StudentID    string     `gorm:"type:uuid;not null;index"`
	// Универсальные значения:
	Scale     string    `gorm:"not null;index"` // дублируем из Assessment для CHECK (passfail|five|hundred)
	ValueNum  *float64  // для five/hundred (five: 2..5; hundred: 0..100)
	ValuePass *bool     // для passfail (true/false)
	GradedBy  string    `gorm:"type:text;not null"`
	GradedAt  time.Time `gorm:"not null;default:now();index"`

	_  struct{} `gorm:"uniqueIndex:uniq_assessment_student,priority:1"`
	__ struct{} `gorm:"uniqueIndex:uniq_assessment_student,priority:2"`
}
