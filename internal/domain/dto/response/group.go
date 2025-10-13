package response

type Group struct {
	ID        string `json:"id"`
	ProgramID string `json:"programId"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Year      int    `json:"year"`
}
