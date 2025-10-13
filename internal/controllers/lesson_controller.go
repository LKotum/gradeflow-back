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

type LessonController struct {
	DB            *gorm.DB
	Cfg           config.Config
	LessonSvc     service.LessonService
	AttendanceSvc service.AttendanceService
}

func NewLessonController(db *gorm.DB, cfg config.Config, lessonSvc service.LessonService, attendanceSvc service.AttendanceService) *LessonController {
	return &LessonController{DB: db, Cfg: cfg, LessonSvc: lessonSvc, AttendanceSvc: attendanceSvc}
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
	lesson, err := h.LessonSvc.Create(in)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toLessonResp(*lesson))
}

// @Summary List lessons
// @Tags lessons
// @Produce json
// @Param courseId query string false "filter by course"
// @Param sessionId query string false "filter by academic session"
// @Param from query string false "starts/ends window from (RFC3339)"
// @Param to query string false "starts/ends window to (RFC3339)"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.LessonList
// @Router /lessons [get]
func (h *LessonController) list(c *gin.Context) {
	var qin req.ListLessonQuery
	_ = c.ShouldBindQuery(&qin)
	lessons, total, appliedLimit, appliedOffset, err := h.LessonSvc.List(qin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Lesson, 0, len(lessons))
	for _, d := range lessons {
		items = append(items, toLessonResp(d))
	}
	c.JSON(http.StatusOK, resp.List[resp.Lesson]{Items: items, Page: resp.Page{Limit: appliedLimit, Offset: appliedOffset, Total: total}})
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
	lesson, err := h.LessonSvc.Get(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, toLessonResp(*lesson))
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
	lesson, err := h.LessonSvc.Update(id, in)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, toLessonResp(*lesson))
}

// @Summary Delete lesson
// @Tags lessons
// @Produce json
// @Param id path string true "lesson id"
// @Success 200 {object} map[string]bool
// @Router /lessons/{id} [delete]
func (h *LessonController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.LessonSvc.Delete(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
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
    qin.LessonID = lessonID
    records, total, appliedLimit, appliedOffset, err := h.AttendanceSvc.List(qin)
    if err != nil {
        c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
        return
    }
    items := make([]resp.Attendance, 0, len(records))
    for _, a := range records {
        items = append(items, toAttendanceResp(a))
    }
    c.JSON(http.StatusOK, resp.List[resp.Attendance]{Items: items, Page: resp.Page{Limit: appliedLimit, Offset: appliedOffset, Total: total}})
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
    var markedBy *string
    if v, ok := c.Get("user"); ok {
        if u, ok2 := v.(*m.User); ok2 {
            id := u.ID
            markedBy = &id
        }
    }
	records, err := h.AttendanceSvc.BulkUpsert(lessonID, entries, markedBy)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
    items := make([]resp.Attendance, 0, len(records))
    for _, a := range records {
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
