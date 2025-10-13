package response

// Swagger list wrappers to avoid generics in annotations

type DepartmentList struct {
	Items []Department `json:"items"`
	Page  Page         `json:"page"`
}
type ProgramList struct {
	Items []Program `json:"items"`
	Page  Page      `json:"page"`
}
type GroupList struct {
	Items []Group `json:"items"`
	Page  Page    `json:"page"`
}
type SubjectList struct {
	Items []Subject `json:"items"`
	Page  Page      `json:"page"`
}
type StudentList struct {
	Items []Student `json:"items"`
	Page  Page      `json:"page"`
}
type CourseList struct {
	Items []Course `json:"items"`
	Page  Page     `json:"page"`
}
type AcademicSessionList struct {
	Items []AcademicSession `json:"items"`
	Page  Page              `json:"page"`
}
type LessonList struct {
	Items []Lesson `json:"items"`
	Page  Page     `json:"page"`
}
type AttendanceList struct {
	Items []Attendance `json:"items"`
	Page  Page         `json:"page"`
}
type AssessmentList struct {
	Items []Assessment `json:"items"`
	Page  Page         `json:"page"`
}
type AssessmentGradeList struct {
	Items []AssessmentGrade `json:"items"`
	Page  Page              `json:"page"`
}
type EnrollmentList struct {
	Items []Enrollment `json:"items"`
	Page  Page         `json:"page"`
}
type TeacherList struct {
	Items []Teacher `json:"items"`
	Page  Page      `json:"page"`
}
type StaffList struct {
	Items []Staff `json:"items"`
	Page  Page    `json:"page"`
}
type AdminList struct {
	Items []Admin `json:"items"`
	Page  Page    `json:"page"`
}
type PracticeList struct {
	Items []Practice `json:"items"`
	Page  Page       `json:"page"`
}
type PracticeEnrollmentList struct {
	Items []PracticeEnrollment `json:"items"`
	Page  Page                 `json:"page"`
}
type ExamSessionList struct {
	Items []ExamSession `json:"items"`
	Page  Page          `json:"page"`
}
type ExamAttemptList struct {
	Items []ExamAttempt `json:"items"`
	Page  Page          `json:"page"`
}
type CreditList struct {
	Items []Credit `json:"items"`
	Page  Page     `json:"page"`
}
