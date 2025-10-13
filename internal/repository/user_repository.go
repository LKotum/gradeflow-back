package repository

import (
	m "gradeflow/internal/domain/models"

	"gorm.io/gorm"
)

type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(u *m.User) error { return r.db.Create(u).Error }
func (r *UserRepository) ByEmail(email string) (*m.User, error) {
	var u m.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
func (r *UserRepository) ByID(id string) (*m.User, error) {
	var u m.User
	if err := r.db.First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
func (r *UserRepository) UpdateFields(id string, fields map[string]any) error {
	return r.db.Model(&m.User{}).Where("id = ?", id).Updates(fields).Error
}
