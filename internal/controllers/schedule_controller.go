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

type ScheduleController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewScheduleController(db *gorm.DB, cfg config.Config) *ScheduleController {
	return &ScheduleController{DB: db, Cfg: cfg}
}

// RegisterRoutes registers schedule endpoints
func (h *ScheduleController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/schedule")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.GET("/student", h.studentSchedule)
}

// @Summary Get student schedule
// @Tags schedule
// @Produce json
// @Param studentId query string true "student id"
// @Param from query string false "RFC3339 from time"
// @Param to query string false "RFC3339 to time"
// @Success 200 {object} response.List[response.ScheduleItem]
// @Router /schedule/student [get]
func (h *ScheduleController) studentSchedule(c *gin.Context) {
	studentID := c.Query("studentId")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, resp.Error{Error: "studentId required"})
		return
	}
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
	// find courses where student is enrolled
	var enrollments []m.Enrollment
	if err := h.DB.Where("student_id = ?", studentID).Find(&enrollments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if len(enrollments) == 0 {
		c.JSON(http.StatusOK, resp.List[resp.ScheduleItem]{Items: []resp.ScheduleItem{}, Page: resp.Page{Limit: 0, Offset: 0, Total: 0}})
		return
	}
	courseIDs := make([]string, 0, len(enrollments))
	for _, e := range enrollments {
		courseIDs = append(courseIDs, e.CourseID)
	}
	// fetch lessons for these courses within optional range
	q := h.DB.Model(&m.Lesson{}).Where("course_id IN ?", courseIDs)
	if fromPtr != nil {
		q = q.Where("ends_at >= ?", *fromPtr)
	}
	if toPtr != nil {
		q = q.Where("starts_at <= ?", *toPtr)
	}
	var lessons []m.Lesson
	if err := q.Order("starts_at asc").Find(&lessons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	// get course titles
	var courses []m.Course
	if err := h.DB.Where("id IN ?", courseIDs).Find(&courses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	titleByID := map[string]string{}
	for _, cr := range courses {
		titleByID[cr.ID] = cr.Title
	}
	items := make([]resp.ScheduleItem, 0, len(lessons))
	for _, l := range lessons {
		items = append(items, resp.ScheduleItem{
			LessonID:    l.ID,
			CourseID:    l.CourseID,
			CourseTitle: titleByID[l.CourseID],
			StartsAt:    l.StartsAt,
			EndsAt:      l.EndsAt,
			Room:        l.Room,
			Kind:        l.Kind,
		})
	}
	c.JSON(http.StatusOK, resp.List[resp.ScheduleItem]{Items: items, Page: resp.Page{Limit: len(items), Offset: 0, Total: int64(len(items))}})
}
