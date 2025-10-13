package response

import "time"

type AcademicSession struct {
	ID       string    `json:"id"`
	Code     string    `json:"code"`
	Kind     string    `json:"kind"`
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
}
