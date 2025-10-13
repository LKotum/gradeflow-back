package response

import "time"

type Lesson struct {
	ID       string    `json:"id"`
	CourseID string    `json:"courseId"`
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
	Room     string    `json:"room"`
	Kind     string    `json:"kind"`
}
