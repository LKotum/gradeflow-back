package models

// Студент (с индивидуальным номером)
type Student struct {
	Base
	IndividualNumber string  `gorm:"uniqueIndex;not null"`
	FullName         string  `gorm:"not null;index"`
	GroupID          *string `gorm:"type:uuid;index"`
	Group            *Group  `gorm:"foreignKey:GroupID;constraint:OnDelete:SET NULL"`
	UserID           *string `gorm:"type:uuid;uniqueIndex"`
	User             *User   `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}
