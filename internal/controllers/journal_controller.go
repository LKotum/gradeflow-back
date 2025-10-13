package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gradeflow/internal/config"
	resp "gradeflow/internal/domain/dto/response"
	m "gradeflow/internal/domain/models"
	"gradeflow/pkg/middleware"
)

type JournalController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewJournalController(db *gorm.DB, cfg config.Config) *JournalController {
	return &JournalController{DB: db, Cfg: cfg}
}

func (h *JournalController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.GET("/students/:id/journal", h.studentJournal)
}

// @Summary Get student journal (attendance and grades)
// @Tags journal
// @Produce json
// @Param id path string true "student id"
// @Param courseId query string false "optional course filter"
// @Param from query string false "RFC3339 from time"
// @Param to query string false "RFC3339 to time"
// @Success 200 {object} response.StudentJournal
// @Router /students/{id}/journal [get]
func (h *JournalController) studentJournal(c *gin.Context) {
	sid := c.Param("id")
	if sid == "" {
		c.JSON(http.StatusBadRequest, resp.Error{Error: "missing id"})
		return
	}
	courseID := c.Query("courseId")
	var fromPtr, toPtr *time.Time
	if s := c.Query("from"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			fromPtr = &t
		}
	}
	if s := c.Query("to"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			toPtr = &t
		}
	}
	// Attendance
	aq := h.DB.Model(&m.Attendance{}).Where("student_id = ?", sid)
	if courseID != "" {
		// attendance references lessons; need join to lessons to filter by course
		aq = aq.Joins("JOIN lessons ON lessons.id = attendances.lesson_id").Where("lessons.course_id = ?", courseID)
	}
	if fromPtr != nil {
		aq = aq.Where("marked_at >= ?", *fromPtr)
	}
	if toPtr != nil {
		aq = aq.Where("marked_at <= ?", *toPtr)
	}
	var attends []m.Attendance
	if err := aq.Find(&attends).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	aItems := make([]resp.JournalAttendanceEntry, 0, len(attends))
	for _, a := range attends {
		aItems = append(aItems, resp.JournalAttendanceEntry{LessonID: a.LessonID, Status: a.Status})
	}

	// Grades
	gq := h.DB.Model(&m.AssessmentGrade{}).Where("student_id = ?", sid)
	if courseID != "" {
		gq = gq.Joins("JOIN assessments ON assessments.id = assessment_grades.assessment_id").Where("assessments.course_id = ?", courseID)
	}
	if fromPtr != nil {
		gq = gq.Where("graded_at >= ?", *fromPtr)
	}
	if toPtr != nil {
		gq = gq.Where("graded_at <= ?", *toPtr)
	}
	var grades []m.AssessmentGrade
	if err := gq.Find(&grades).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	gItems := make([]resp.JournalGradeEntry, 0, len(grades))
	for _, g0 := range grades {
		// type is on assessment; fetch minimal types in batch would be better; for now omit Type field if not available
		gItems = append(gItems, resp.JournalGradeEntry{AssessmentID: g0.ID, Scale: g0.Scale, ValueNum: g0.ValueNum, ValuePass: g0.ValuePass})
	}

	var cidPtr *string
	if courseID != "" {
		cidPtr = &courseID
	}
	c.JSON(http.StatusOK, resp.StudentJournal{StudentID: sid, CourseID: cidPtr, Attendance: aItems, Grades: gItems})
}
