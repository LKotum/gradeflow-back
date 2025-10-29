package response

// PaginatedUserProfiles is used for Swagger documentation to describe paginated user profiles.
type PaginatedUserProfiles struct {
	Data []UserProfile `json:"data"`
	Meta PageMeta      `json:"meta"`
}

// PaginatedGroupSummaries is used for Swagger documentation to describe paginated groups.
type PaginatedGroupSummaries struct {
	Data []GroupSummary `json:"data"`
	Meta PageMeta       `json:"meta"`
}

// PaginatedSubjectSummaries is used for Swagger documentation to describe paginated subjects.
type PaginatedSubjectSummaries struct {
	Data []SubjectSummary `json:"data"`
	Meta PageMeta         `json:"meta"`
}

// PaginatedStudentSubjects is used for Swagger documentation to describe paginated student subjects.
type PaginatedStudentSubjects struct {
	Data []StudentSubjectGrade `json:"data"`
	Meta PageMeta              `json:"meta"`
}
