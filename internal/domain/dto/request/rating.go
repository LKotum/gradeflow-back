package request

type GroupRatingQuery struct {
	CourseID  string `form:"courseId"`
	SessionID string `form:"sessionId"`
	Limit     int    `form:"limit"`
}

type StudentRatingQuery struct {
	CourseID  string `form:"courseId"`
	SessionID string `form:"sessionId"`
	GroupID   string `form:"groupId"`
	Limit     int    `form:"limit"`
}

type LimitQuery struct {
	Limit int `form:"limit"`
}
