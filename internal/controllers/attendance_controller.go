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
	"gradeflow/internal/service"
	"gradeflow/pkg/middleware"
)

type AttendanceController struct {
	DB      *gorm.DB
	Cfg     config.Config
	Service service.AttendanceService
}

func NewAttendanceController(db *gorm.DB, cfg config.Config, svc service.AttendanceService) *AttendanceController {
	return &AttendanceController{DB: db, Cfg: cfg, Service: svc}
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
	var markedBy *string
	if v, ok := c.Get("user"); ok {
		if u, ok2 := v.(*m.User); ok2 {
			id := u.ID
			markedBy = &id
		}
	}
	d, err := h.Service.Create(in, markedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, attendanceToResp(*d))
}

// @Summary List attendance
// @Tags attendance
// @Produce json
// @Param lessonId query string false "filter by lesson"
// @Param studentId query string false "filter by student"
// @Param courseId query string false "filter by course"
// @Param sessionId query string false "filter by academic session"
// @Param from query string false "marked from (RFC3339)"
// @Param to query string false "marked to (RFC3339)"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.AttendanceList
// @Router /attendance [get]
func (h *AttendanceController) list(c *gin.Context) {
	var qin req.ListAttendanceQuery
	_ = c.ShouldBindQuery(&qin)
	records, total, appliedLimit, appliedOffset, err := h.Service.List(qin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Attendance, 0, len(records))
	for _, d := range records {
		items = append(items, attendanceToResp(d))
	}
	c.JSON(http.StatusOK, resp.List[resp.Attendance]{Items: items, Page: resp.Page{Limit: appliedLimit, Offset: appliedOffset, Total: total}})
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
	att, err := h.Service.Get(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, attendanceToResp(*att))
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
	var markedBy *string
	if v, ok := c.Get("user"); ok {
		if u, ok2 := v.(*m.User); ok2 {
			id := u.ID
			markedBy = &id
		}
	}
	att, err := h.Service.Update(id, in, markedBy)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, attendanceToResp(*att))
}

// @Summary Delete attendance
// @Tags attendance
// @Produce json
// @Param id path string true "attendance id"
// @Success 200 {object} map[string]bool
// @Router /attendance/{id} [delete]
func (h *AttendanceController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
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
