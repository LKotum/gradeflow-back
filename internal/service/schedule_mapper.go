package service

import (
	"gradeflow/internal/domain/models"
	respdto "gradeflow/internal/domain/dto/response"
	"github.com/google/uuid"
)

func buildScheduleEntries(sessions []models.ClassSession) []respdto.ScheduleEntry {
	entries := make([]respdto.ScheduleEntry, 0, len(sessions))
	for _, session := range sessions {
		entry := respdto.ScheduleEntry{
			Session: respdto.SessionSummary{
				ID:        session.ID.String(),
				StartsAt:  session.StartsAt,
				EndsAt:    session.EndsAt,
				Topic:     session.Topic,
				SubjectID: session.SubjectID.String(),
				GroupID:   session.GroupID.String(),
			},
			Subject: respdto.SubjectSummary{
				ID:          session.SubjectID.String(),
				Code:        session.Subject.Code,
				Name:        session.Subject.Name,
				Description: session.Subject.Description,
			},
			Group: respdto.GroupSummary{
				ID:          session.GroupID.String(),
				Name:        session.Group.Name,
				Description: session.Group.Description,
			},
		}

		if session.Teacher.ID != uuid.Nil || session.Teacher.FirstName != "" || session.Teacher.LastName != "" {
			entry.Teacher = &respdto.UserProfile{
				ID:         session.Teacher.ID.String(),
				FirstName:  session.Teacher.FirstName,
				LastName:   session.Teacher.LastName,
				MiddleName: session.Teacher.MiddleName,
				Email:      session.Teacher.Email,
				INS:        session.Teacher.INS,
				AvatarURL:  session.Teacher.AvatarURL,
				Role:       string(session.Teacher.Role),
			}
		}

		entries = append(entries, entry)
	}
	return entries
}
