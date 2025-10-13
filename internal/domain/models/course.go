package models

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
	TeacherID         *string         `gorm:"type:uuid;index"`
	Teacher           *Teacher        `gorm:"foreignKey:TeacherID;constraint:OnDelete:SET NULL"`
	Room              string          `gorm:"index"`
}
