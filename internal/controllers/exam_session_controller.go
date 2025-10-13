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

type ExamSessionController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewExamSessionController(db *gorm.DB, cfg config.Config) *ExamSessionController {
	return &ExamSessionController{DB: db, Cfg: cfg}
}

func (h *ExamSessionController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/exam-sessions")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)
}

// @Summary Create exam session
// @Tags exam-sessions
// @Accept json
// @Produce json
// @Param input body request.CreateExamSession true "exam session"
// @Success 201 {object} response.ExamSession
// @Failure 400 {object} response.Error
// @Router /exam-sessions [post]
func (h *ExamSessionController) create(c *gin.Context) {
	var in req.CreateExamSession
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	es := m.ExamSession{AcademicSessionID: in.AcademicSessionID, Name: in.Name}
	// parse times if provided in RFC3339
	if in.StartsAt != nil {
		es.StartsAt = parseTimePtr(*in.StartsAt)
	}
	if in.EndsAt != nil {
		es.EndsAt = parseTimePtr(*in.EndsAt)
	}
	if err := h.DB.Create(&es).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toExamSessionResp(es))
}

// @Summary List exam sessions
// @Tags exam-sessions
// @Produce json
// @Param academicSessionId query string false "filter by academic session"
// @Param q query string false "search by name"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.ExamSessionList
// @Router /exam-sessions [get]
func (h *ExamSessionController) list(c *gin.Context) {
	var qin req.ListExamSessionQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.ExamSession{})
	if v := qin.AcademicSessionID; v != "" {
		base = base.Where("academic_session_id = ?", v)
	}
	if v := qin.Q; v != "" {
		base = base.Where("name ILIKE ?", "%"+v+"%")
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if qin.Limit <= 0 {
		qin.Limit = 50
	}
	if qin.Limit > 200 {
		qin.Limit = 200
	}
	if qin.Offset < 0 {
		qin.Offset = 0
	}
	var ll []m.ExamSession
	if err := base.Order("starts_at asc nulls last").Limit(qin.Limit).Offset(qin.Offset).Find(&ll).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.ExamSession, 0, len(ll))
	for _, es := range ll {
		items = append(items, toExamSessionResp(es))
	}
	c.JSON(http.StatusOK, resp.List[resp.ExamSession]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get exam session
// @Tags exam-sessions
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} response.ExamSession
// @Failure 404 {object} response.Error
// @Router /exam-sessions/{id} [get]
func (h *ExamSessionController) get(c *gin.Context) {
	id := c.Param("id")
	var es m.ExamSession
	if err := h.DB.First(&es, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toExamSessionResp(es))
}

// @Summary Update exam session
// @Tags exam-sessions
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param input body request.UpdateExamSession true "exam session"
// @Success 200 {object} response.ExamSession
// @Failure 400 {object} response.Error
// @Router /exam-sessions/{id} [put]
func (h *ExamSessionController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateExamSession
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var es m.ExamSession
	if err := h.DB.First(&es, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.Name != nil {
		es.Name = *in.Name
	}
	if in.StartsAt != nil {
		es.StartsAt = parseTimePtr(*in.StartsAt)
	}
	if in.EndsAt != nil {
		es.EndsAt = parseTimePtr(*in.EndsAt)
	}
	if err := h.DB.Save(&es).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toExamSessionResp(es))
}

// @Summary Delete exam session
// @Tags exam-sessions
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} map[string]bool
// @Router /exam-sessions/{id} [delete]
func (h *ExamSessionController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.ExamSession{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func parseTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	return nil
}

func toExamSessionResp(es m.ExamSession) resp.ExamSession {
	var starts, ends *string
	if es.StartsAt != nil {
		s := es.StartsAt.Format(time.RFC3339)
		starts = &s
	}
	if es.EndsAt != nil {
		s := es.EndsAt.Format(time.RFC3339)
		ends = &s
	}
	return resp.ExamSession{ID: es.ID, AcademicSessionID: es.AcademicSessionID, Name: es.Name, StartsAt: starts, EndsAt: ends}
}
