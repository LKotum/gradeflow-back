package response

type Student struct {
	ID               string  `json:"id"`
	IndividualNumber string  `json:"individualNumber"`
	FullName         string  `json:"fullName"`
	GroupID          *string `json:"groupId"`
	StartYear        *int    `json:"startYear,omitempty"`
	EndYear          *int    `json:"endYear,omitempty"`
}
