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

type ProgramController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewProgramController(db *gorm.DB, cfg config.Config) *ProgramController {
	return &ProgramController{DB: db, Cfg: cfg}
}

func (h *ProgramController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/programs")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)
}

// @Summary Create program
// @Tags programs
// @Accept json
// @Produce json
// @Param input body request.CreateProgram true "program"
// @Success 201 {object} response.Program
// @Router /programs [post]
func (h *ProgramController) create(c *gin.Context) {
	var in req.CreateProgram
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	p := m.Program{DepartmentID: in.DepartmentID, Code: in.Code, Name: in.Name}
	if err := h.DB.Create(&p).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toProgramResp(p))
}

// @Summary List programs
// @Tags programs
// @Produce json
// @Success 200 {object} response.ProgramList
// @Router /programs [get]
func (h *ProgramController) list(c *gin.Context) {
	var qin req.ListProgramQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Program{})
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
	var pp []m.Program
	if err := base.Order("code asc").Limit(qin.Limit).Offset(qin.Offset).Find(&pp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Program, 0, len(pp))
	for _, p := range pp {
		items = append(items, toProgramResp(p))
	}
	c.JSON(http.StatusOK, resp.List[resp.Program]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get program
// @Tags programs
// @Produce json
// @Param id path string true "program id"
// @Success 200 {object} response.Program
// @Router /programs/{id} [get]
func (h *ProgramController) get(c *gin.Context) {
	id := c.Param("id")
	var p m.Program
	if err := h.DB.First(&p, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toProgramResp(p))
}

// @Summary Update program
// @Tags programs
// @Accept json
// @Produce json
// @Param id path string true "program id"
// @Param input body request.UpdateProgram true "program"
// @Success 200 {object} response.Program
// @Router /programs/{id} [put]
func (h *ProgramController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateProgram
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var p m.Program
	if err := h.DB.First(&p, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.DepartmentID != "" {
		p.DepartmentID = in.DepartmentID
	}
	if in.Code != "" {
		p.Code = in.Code
	}
	if in.Name != "" {
		p.Name = in.Name
	}
	if err := h.DB.Save(&p).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toProgramResp(p))
}

// @Summary Delete program
// @Tags programs
// @Produce json
// @Param id path string true "program id"
// @Success 200 {object} map[string]bool
// @Router /programs/{id} [delete]
func (h *ProgramController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Program{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func toProgramResp(p m.Program) resp.Program {
	return resp.Program{ID: p.ID, DepartmentID: p.DepartmentID, Code: p.Code, Name: p.Name}
}
