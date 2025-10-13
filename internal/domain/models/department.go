package models

// Кафедра
type Department struct {
	Base
	Code string `gorm:"uniqueIndex;not null"`
	Name string `gorm:"not null;index"`
}
