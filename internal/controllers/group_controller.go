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

type GroupController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewGroupController(db *gorm.DB, cfg config.Config) *GroupController {
	return &GroupController{DB: db, Cfg: cfg}
}

func (h *GroupController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/groups")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)
}

// @Summary Create group
// @Tags groups
// @Accept json
// @Produce json
// @Param input body request.CreateGroup true "group"
// @Success 201 {object} response.Group
// @Router /groups [post]
func (h *GroupController) create(c *gin.Context) {
	var in req.CreateGroup
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	g := m.Group{ProgramID: in.ProgramID, Code: in.Code, Name: in.Name, Year: in.Year}
	if err := h.DB.Create(&g).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toGroupResp(g))
}

// @Summary List groups
// @Tags groups
// @Produce json
// @Success 200 {object} response.GroupList
// @Router /groups [get]
func (h *GroupController) list(c *gin.Context) {
	var qin req.ListGroupQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Group{})
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
	var gg []m.Group
	if err := base.Order("code asc").Limit(qin.Limit).Offset(qin.Offset).Find(&gg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Group, 0, len(gg))
	for _, g := range gg {
		items = append(items, toGroupResp(g))
	}
	c.JSON(http.StatusOK, resp.List[resp.Group]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get group
// @Tags groups
// @Produce json
// @Param id path string true "group id"
// @Success 200 {object} response.Group
// @Router /groups/{id} [get]
func (h *GroupController) get(c *gin.Context) {
	id := c.Param("id")
	var g m.Group
	if err := h.DB.First(&g, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toGroupResp(g))
}

// @Summary Update group
// @Tags groups
// @Accept json
// @Produce json
// @Param id path string true "group id"
// @Param input body request.UpdateGroup true "group"
// @Success 200 {object} response.Group
// @Router /groups/{id} [put]
func (h *GroupController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateGroup
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var g m.Group
	if err := h.DB.First(&g, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.ProgramID != "" {
		g.ProgramID = in.ProgramID
	}
	if in.Code != "" {
		g.Code = in.Code
	}
	if in.Name != "" {
		g.Name = in.Name
	}
	if in.Year != 0 {
		g.Year = in.Year
	}
	if err := h.DB.Save(&g).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toGroupResp(g))
}

// @Summary Delete group
// @Tags groups
// @Produce json
// @Param id path string true "group id"
// @Success 200 {object} map[string]bool
// @Router /groups/{id} [delete]
func (h *GroupController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Group{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func toGroupResp(g m.Group) resp.Group {
	return resp.Group{ID: g.ID, ProgramID: g.ProgramID, Code: g.Code, Name: g.Name, Year: g.Year}
}
