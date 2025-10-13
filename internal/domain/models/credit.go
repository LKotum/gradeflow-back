package models

// Перезачет/зачет по предмету (offset/credit)
type Credit struct {
	Base
	CourseID     string  `gorm:"type:uuid;not null;index"`
	Course       Course  `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
	StudentID    string  `gorm:"type:uuid;not null;index"`
	Student      Student `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	Type         string  `gorm:"not null;index"` // credit|offset|transfer
	Reason       string  `gorm:"type:text"`
	ApprovedByID *string `gorm:"type:uuid;index"` // StaffMember or Admin by policy
}
