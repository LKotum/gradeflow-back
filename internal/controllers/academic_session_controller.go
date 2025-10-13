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

type AcademicSessionController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewAcademicSessionController(db *gorm.DB, cfg config.Config) *AcademicSessionController {
	return &AcademicSessionController{DB: db, Cfg: cfg}
}

func (h *AcademicSessionController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/academic-sessions")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)
}

// @Summary Create academic session
// @Tags academic-sessions
// @Accept json
// @Produce json
// @Param input body request.CreateAcademicSession true "session"
// @Success 201 {object} response.AcademicSession
// @Failure 400 {object} response.Error
// @Router /academic-sessions [post]
func (h *AcademicSessionController) create(c *gin.Context) {
	var in req.CreateAcademicSession
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	d := m.AcademicSession{Code: in.Code, Kind: in.Kind, StartsAt: in.StartsAt, EndsAt: in.EndsAt}
	if err := h.DB.Create(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toAcademicSessionResp(d))
}

// @Summary List academic sessions
// @Tags academic-sessions
// @Produce json
// @Param from query string false "window from (RFC3339)"
// @Param to query string false "window to (RFC3339)"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.AcademicSessionList
// @Router /academic-sessions [get]
func (h *AcademicSessionController) list(c *gin.Context) {
	var qin req.ListAcademicSessionQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.AcademicSession{})
	if qin.From != "" {
		if t, err := time.Parse(time.RFC3339, qin.From); err == nil {
			base = base.Where("ends_at >= ?", t)
		}
	}
	if qin.To != "" {
		if t, err := time.Parse(time.RFC3339, qin.To); err == nil {
			base = base.Where("starts_at <= ?", t)
		}
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if qin.Limit <= 0 {
		qin.Limit = 200
	}
	if qin.Limit > 500 {
		qin.Limit = 500
	}
	if qin.Offset < 0 {
		qin.Offset = 0
	}
	var dd []m.AcademicSession
	if err := base.Order("starts_at asc").Limit(qin.Limit).Offset(qin.Offset).Find(&dd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.AcademicSession, 0, len(dd))
	for _, d := range dd {
		items = append(items, toAcademicSessionResp(d))
	}
	c.JSON(http.StatusOK, resp.List[resp.AcademicSession]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get academic session
// @Tags academic-sessions
// @Produce json
// @Param id path string true "session id"
// @Success 200 {object} response.AcademicSession
// @Failure 404 {object} response.Error
// @Router /academic-sessions/{id} [get]
func (h *AcademicSessionController) get(c *gin.Context) {
	id := c.Param("id")
	var d m.AcademicSession
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toAcademicSessionResp(d))
}

// @Summary Update academic session
// @Tags academic-sessions
// @Accept json
// @Produce json
// @Param id path string true "session id"
// @Param input body request.UpdateAcademicSession true "session"
// @Success 200 {object} response.AcademicSession
// @Failure 400 {object} response.Error
// @Router /academic-sessions/{id} [put]
func (h *AcademicSessionController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateAcademicSession
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var d m.AcademicSession
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.Code != "" {
		d.Code = in.Code
	}
	if in.Kind != "" {
		d.Kind = in.Kind
	}
	if in.StartsAt != nil {
		d.StartsAt = *in.StartsAt
	}
	if in.EndsAt != nil {
		d.EndsAt = *in.EndsAt
	}
	if err := h.DB.Save(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toAcademicSessionResp(d))
}

// @Summary Delete academic session
// @Tags academic-sessions
// @Produce json
// @Param id path string true "session id"
// @Success 200 {object} response.OK
// @Router /academic-sessions/{id} [delete]
func (h *AcademicSessionController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.AcademicSession{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.OK{OK: true})
}

func toAcademicSessionResp(a m.AcademicSession) resp.AcademicSession {
	return resp.AcademicSession{ID: a.ID, Code: a.Code, Kind: a.Kind, StartsAt: a.StartsAt, EndsAt: a.EndsAt}
}
