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

type AttendanceController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewAttendanceController(db *gorm.DB, cfg config.Config) *AttendanceController {
	return &AttendanceController{DB: db, Cfg: cfg}
}

func (h *AttendanceController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/attendance")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.StaffOrTeacher(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.StaffOrTeacher(), h.update)
	g.DELETE(":id", middleware.StaffOrTeacher(), h.delete)
}

// @Summary Create attendance
// @Tags attendance
// @Accept json
// @Produce json
// @Param input body request.CreateAttendance true "attendance"
// @Success 201 {object} response.Attendance
// @Failure 400 {object} response.Error
// @Router /attendance [post]
func (h *AttendanceController) create(c *gin.Context) {
	var in req.CreateAttendance
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	if in.Status != "present" && in.Status != "absent" && in.Status != "late" {
		c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid status"})
		return
	}
	var markedBy string
	if v, ok := c.Get("user"); ok {
		if u, ok2 := v.(*m.User); ok2 {
			markedBy = u.ID
		}
	}
	d := m.Attendance{LessonID: in.LessonID, StudentID: in.StudentID, Status: in.Status, MarkedBy: markedBy, MarkedAt: time.Now()}
	if err := h.DB.Create(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, attendanceToResp(d))
}

// @Summary List attendance
// @Tags attendance
// @Produce json
// @Success 200 {object} response.AttendanceList
// @Router /attendance [get]
func (h *AttendanceController) list(c *gin.Context) {
	var qin req.ListAttendanceQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Attendance{})
	if v := qin.LessonID; v != "" {
		base = base.Where("lesson_id = ?", v)
	}
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
	var dd []m.Attendance
	if err := base.Limit(qin.Limit).Offset(qin.Offset).Find(&dd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Attendance, 0, len(dd))
	for _, d := range dd {
		items = append(items, attendanceToResp(d))
	}
	c.JSON(http.StatusOK, resp.List[resp.Attendance]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get attendance
// @Tags attendance
// @Produce json
// @Param id path string true "attendance id"
// @Success 200 {object} response.Attendance
// @Failure 404 {object} response.Error
// @Router /attendance/{id} [get]
func (h *AttendanceController) get(c *gin.Context) {
	id := c.Param("id")
	var d m.Attendance
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, attendanceToResp(d))
}

// @Summary Update attendance
// @Tags attendance
// @Accept json
// @Produce json
// @Param id path string true "attendance id"
// @Param input body request.UpdateAttendance true "attendance"
// @Success 200 {object} response.Attendance
// @Failure 400 {object} response.Error
// @Router /attendance/{id} [put]
func (h *AttendanceController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateAttendance
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var d m.Attendance
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.Status != "" {
		if in.Status != "present" && in.Status != "absent" && in.Status != "late" {
			c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid status"})
			return
		}
		d.Status = in.Status
	}
	if err := h.DB.Save(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, attendanceToResp(d))
}

// @Summary Delete attendance
// @Tags attendance
// @Produce json
// @Param id path string true "attendance id"
// @Success 200 {object} map[string]bool
// @Router /attendance/{id} [delete]
func (h *AttendanceController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Attendance{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.OK{OK: true})
}

func attendanceToResp(a m.Attendance) resp.Attendance {
	var mb *string
	if a.MarkedBy != "" {
		mb = &a.MarkedBy
	}
	return resp.Attendance{ID: a.ID, LessonID: a.LessonID, StudentID: a.StudentID, Status: a.Status, MarkedBy: mb, MarkedAt: a.MarkedAt.Format(time.RFC3339)}
}
