package service

import (
	"fmt"
	"strings"

	"gradeflow/internal/domain/models"
)

func stringPtr(s string) *string {
	return &s
}

func studentIndexOrDefault(user *models.User) string {
	if user == nil {
		return ""
	}
	if user.Student != nil {
		if trimmed := strings.TrimSpace(user.Student.Index); trimmed != "" {
			return trimmed
		}
	}
	if user.INS != nil {
		if trimmed := strings.TrimSpace(*user.INS); trimmed != "" {
			return fmt.Sprintf("ST-%s", trimmed)
		}
	}
	return ""
}
