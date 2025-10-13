package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gradeflow/internal/config"
	req "gradeflow/internal/domain/dto/request"
	resp "gradeflow/internal/domain/dto/response"
	m "gradeflow/internal/domain/models"
	"gradeflow/pkg/middleware"
)

type SubjectController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewSubjectController(db *gorm.DB, cfg config.Config) *SubjectController {
	return &SubjectController{DB: db, Cfg: cfg}
}

func (h *SubjectController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/subjects")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)
}

// @Summary Create subject
// @Tags subjects
// @Accept json
// @Produce json
// @Param input body request.CreateSubject true "subject"
// @Success 201 {object} response.Subject
// @Router /subjects [post]
func (h *SubjectController) create(c *gin.Context) {
	var in req.CreateSubject
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	s := m.Subject{DepartmentID: in.DepartmentID, Code: in.Code, Title: in.Title, Credits: in.Credits}
	if err := h.DB.Create(&s).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toSubjectResp(s))
}

// @Summary List subjects
// @Tags subjects
// @Produce json
// @Success 200 {object} response.SubjectList
// @Router /subjects [get]
func (h *SubjectController) list(c *gin.Context) {
	var qin req.ListSubjectQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Subject{})
	var total int64
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if qin.Limit <= 0 {
		qin.Limit = 100
	}
	if qin.Limit > 500 {
		qin.Limit = 500
	}
	if qin.Offset < 0 {
		qin.Offset = 0
	}
	var ss []m.Subject
	if err := base.Order("code asc").Limit(qin.Limit).Offset(qin.Offset).Find(&ss).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Subject, 0, len(ss))
	for _, s := range ss {
		items = append(items, toSubjectResp(s))
	}
	c.JSON(http.StatusOK, resp.List[resp.Subject]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get subject
// @Tags subjects
// @Produce json
// @Param id path string true "subject id"
// @Success 200 {object} response.Subject
// @Router /subjects/{id} [get]
func (h *SubjectController) get(c *gin.Context) {
	id := c.Param("id")
	var s m.Subject
	if err := h.DB.First(&s, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toSubjectResp(s))
}

// @Summary Update subject
// @Tags subjects
// @Accept json
// @Produce json
// @Param id path string true "subject id"
// @Param input body request.UpdateSubject true "subject"
// @Success 200 {object} response.Subject
// @Router /subjects/{id} [put]
func (h *SubjectController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateSubject
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var s m.Subject
	if err := h.DB.First(&s, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.DepartmentID != "" {
		s.DepartmentID = in.DepartmentID
	}
	if in.Code != "" {
		s.Code = in.Code
	}
	if in.Title != "" {
		s.Title = in.Title
	}
	if in.Credits != 0 {
		s.Credits = in.Credits
	}
	if err := h.DB.Save(&s).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toSubjectResp(s))
}

// @Summary Delete subject
// @Tags subjects
// @Produce json
// @Param id path string true "subject id"
// @Success 200 {object} map[string]bool
// @Router /subjects/{id} [delete]
func (h *SubjectController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Subject{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func toSubjectResp(s m.Subject) resp.Subject {
	return resp.Subject{ID: s.ID, DepartmentID: s.DepartmentID, Code: s.Code, Title: s.Title, Credits: s.Credits}
}
