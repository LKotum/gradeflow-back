package service

import (
	"fmt"
	"strings"

	respdto "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/domain/models"
)

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

func userProfileFromModel(u *models.User) respdto.UserProfile {
	if u == nil {
		return respdto.UserProfile{}
	}
	profile := respdto.UserProfile{
		ID:         u.ID.String(),
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		MiddleName: u.MiddleName,
		Email:      u.Email,
		INS:        u.INS,
		AvatarURL:  u.AvatarURL,
		Role:       string(u.Role),
	}
	if u.Student != nil {
		if u.Student.Group != nil {
			grp := u.Student.Group
			profile.Group = &respdto.GroupSummary{
				ID:          grp.ID.String(),
				Name:        grp.Name,
				Description: grp.Description,
			}
		} else if u.Student.GroupID != nil {
			profile.Group = &respdto.GroupSummary{ID: u.Student.GroupID.String()}
		}
	}
	if u.Teacher != nil {
		profile.TeacherTitle = u.Teacher.Title
		profile.TeacherBio = u.Teacher.Bio
	}
	if u.Staff != nil {
		profile.StaffPosition = u.Staff.Position
	}
	return profile
}
