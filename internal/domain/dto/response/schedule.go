package response

import "time"

// ScheduleItem represents a lesson planned for a student along with course title
type ScheduleItem struct {
	LessonID    string    `json:"lessonId"`
	CourseID    string    `json:"courseId"`
	CourseTitle string    `json:"courseTitle"`
	StartsAt    time.Time `json:"startsAt"`
	EndsAt      time.Time `json:"endsAt"`
	Room        string    `json:"room"`
	Kind        string    `json:"kind"`
}
