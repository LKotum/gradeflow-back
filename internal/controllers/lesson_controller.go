package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gradeflow/internal/config"
	req "gradeflow/internal/domain/dto/request"
	resp "gradeflow/internal/domain/dto/response"
	m "gradeflow/internal/domain/models"
	"gradeflow/pkg/middleware"
)

type LessonController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewLessonController(db *gorm.DB, cfg config.Config) *LessonController {
	return &LessonController{DB: db, Cfg: cfg}
}

func (h *LessonController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/lessons")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.StaffOrTeacher(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.StaffOrTeacher(), h.update)
	g.DELETE(":id", middleware.StaffOrTeacher(), h.delete)

	// nested attendance endpoints
	g.GET(":id/attendance", h.listAttendance)
	g.POST(":id/attendance/bulk", middleware.StaffOrTeacher(), h.bulkAttendance)
}

// @Summary Create lesson
// @Tags lessons
// @Accept json
// @Produce json
// @Param input body request.CreateLesson true "lesson"
// @Success 201 {object} response.Lesson
// @Failure 400 {object} response.Error
// @Router /lessons [post]
func (h *LessonController) create(c *gin.Context) {
	var in req.CreateLesson
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	d := m.Lesson{CourseID: in.CourseID, StartsAt: in.StartsAt, EndsAt: in.EndsAt, Room: in.Room, Kind: in.Kind}
	if err := h.DB.Create(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toLessonResp(d))
}

// @Summary List lessons
// @Tags lessons
// @Produce json
// @Success 200 {object} response.LessonList
// @Router /lessons [get]
func (h *LessonController) list(c *gin.Context) {
	var qin req.ListLessonQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Lesson{})
	if v := qin.CourseID; v != "" {
		base = base.Where("course_id = ?", v)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if qin.Limit <= 0 {
		qin.Limit = 200
	}
	if qin.Limit > 1000 {
		qin.Limit = 1000
	}
	if qin.Offset < 0 {
		qin.Offset = 0
	}
	var dd []m.Lesson
	if err := base.Order("starts_at asc").Limit(qin.Limit).Offset(qin.Offset).Find(&dd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Lesson, 0, len(dd))
	for _, d := range dd {
		items = append(items, toLessonResp(d))
	}
	c.JSON(http.StatusOK, resp.List[resp.Lesson]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get lesson
// @Tags lessons
// @Produce json
// @Param id path string true "lesson id"
// @Success 200 {object} response.Lesson
// @Failure 404 {object} response.Error
// @Router /lessons/{id} [get]
func (h *LessonController) get(c *gin.Context) {
	id := c.Param("id")
	var d m.Lesson
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toLessonResp(d))
}

// @Summary Update lesson
// @Tags lessons
// @Accept json
// @Produce json
// @Param id path string true "lesson id"
// @Param input body request.UpdateLesson true "lesson"
// @Success 200 {object} response.Lesson
// @Failure 400 {object} response.Error
// @Router /lessons/{id} [put]
func (h *LessonController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateLesson
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var d m.Lesson
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.CourseID != "" {
		d.CourseID = in.CourseID
	}
	if in.StartsAt != nil {
		d.StartsAt = *in.StartsAt
	}
	if in.EndsAt != nil {
		d.EndsAt = *in.EndsAt
	}
	if in.Room != "" {
		d.Room = in.Room
	}
	if in.Kind != "" {
		d.Kind = in.Kind
	}
	if err := h.DB.Save(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toLessonResp(d))
}

// @Summary Delete lesson
// @Tags lessons
// @Produce json
// @Param id path string true "lesson id"
// @Success 200 {object} map[string]bool
// @Router /lessons/{id} [delete]
func (h *LessonController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Lesson{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// @Summary List attendance for lesson
// @Tags lessons,attendance
// @Produce json
// @Param id path string true "lesson id"
// @Success 200 {object} response.AttendanceList
// @Router /lessons/{id}/attendance [get]
func (h *LessonController) listAttendance(c *gin.Context) {
	lessonID := c.Param("id")
	var qin req.ListAttendanceQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Attendance{}).Where("lesson_id = ?", lessonID)
	if v := qin.StudentID; v != "" {
		base = base.Where("student_id = ?", v)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if qin.Limit <= 0 {
		qin.Limit = 100
	}
	if qin.Limit > 1000 {
		qin.Limit = 1000
	}
	if qin.Offset < 0 {
		qin.Offset = 0
	}
	var aa []m.Attendance
	if err := base.Limit(qin.Limit).Offset(qin.Offset).Find(&aa).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Attendance, 0, len(aa))
	for _, a := range aa {
		items = append(items, toAttendanceResp(a))
	}
	c.JSON(http.StatusOK, resp.List[resp.Attendance]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Bulk upsert attendance for lesson
// @Tags lessons,attendance
// @Accept json
// @Produce json
// @Param id path string true "lesson id"
// @Param input body []request.AttendanceBulkItem true "attendance entries"
// @Success 200 {object} response.AttendanceList
// @Failure 400 {object} response.Error
// @Router /lessons/{id}/attendance/bulk [post]
func (h *LessonController) bulkAttendance(c *gin.Context) {
	lessonID := c.Param("id")
	var entries []req.AttendanceBulkItem
	if err := c.ShouldBindJSON(&entries); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	// who marks
	var markedBy string
	if v, ok := c.Get("user"); ok {
		if u, ok2 := v.(*m.User); ok2 {
			markedBy = u.ID
		}
	}
	now := time.Now()
	// upsert per entry
	for _, e := range entries {
		if e.Status != "present" && e.Status != "absent" && e.Status != "late" {
			c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid status"})
			return
		}
		var rec m.Attendance
		if err := h.DB.Where("lesson_id = ? AND student_id = ?", lessonID, e.StudentID).First(&rec).Error; err == nil {
			rec.Status = e.Status
			if markedBy != "" {
				rec.MarkedBy = markedBy
			}
			rec.MarkedAt = now
			_ = h.DB.Save(&rec).Error
		} else {
			rec = m.Attendance{LessonID: lessonID, StudentID: e.StudentID, Status: e.Status, MarkedBy: markedBy, MarkedAt: now}
			_ = h.DB.Create(&rec).Error
		}
	}
	var out []m.Attendance
	_ = h.DB.Where("lesson_id = ?", lessonID).Find(&out).Error
	items := make([]resp.Attendance, 0, len(out))
	for _, a := range out {
		items = append(items, toAttendanceResp(a))
	}
	c.JSON(http.StatusOK, resp.List[resp.Attendance]{Items: items, Page: resp.Page{Limit: len(items), Offset: 0, Total: int64(len(items))}})
}

func toLessonResp(l m.Lesson) resp.Lesson {
	return resp.Lesson{ID: l.ID, CourseID: l.CourseID, StartsAt: l.StartsAt, EndsAt: l.EndsAt, Room: l.Room, Kind: l.Kind}
}

func toAttendanceResp(a m.Attendance) resp.Attendance {
	var mb *string
	if a.MarkedBy != "" {
		mb = &a.MarkedBy
	}
	return resp.Attendance{ID: a.ID, LessonID: a.LessonID, StudentID: a.StudentID, Status: a.Status, MarkedBy: mb, MarkedAt: a.MarkedAt.Format(time.RFC3339)}
}
