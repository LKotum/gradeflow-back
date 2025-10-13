package repository

import (
	m "gradeflow/internal/domain/models"
	"time"

	"gorm.io/gorm"
)

type TokenRepository struct{ db *gorm.DB }

func NewTokenRepository(db *gorm.DB) *TokenRepository { return &TokenRepository{db: db} }

// RefreshToken
func (r *TokenRepository) CreateRefreshToken(t *m.RefreshToken) error { return r.db.Create(t).Error }
func (r *TokenRepository) FindValidRefreshToken(token string) (*m.RefreshToken, error) {
	var rt m.RefreshToken
	if err := r.db.Where("token = ? and revoked = false and expires_at > ?", token, time.Now()).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}
func (r *TokenRepository) RevokeRefreshToken(rt *m.RefreshToken) error {
	return r.db.Model(rt).Update("revoked", true).Error
}
func (r *TokenRepository) RevokeAllForUser(userID string) error {
	return r.db.Model(&m.RefreshToken{}).Where("user_id = ? and revoked = false", userID).Update("revoked", true).Error
}
