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

type AdminController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewAdminController(db *gorm.DB, cfg config.Config) *AdminController {
	return &AdminController{DB: db, Cfg: cfg}
}

func (h *AdminController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/admins")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.OnlyAdmin(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.OnlyAdmin(), h.update)
	g.DELETE(":id", middleware.OnlyAdmin(), h.delete)
}

// @Summary Create admin
// @Tags admins
// @Accept json
// @Produce json
// @Param input body request.CreateAdmin true "admin"
// @Success 201 {object} response.Admin
// @Failure 400 {object} response.Error
// @Router /admins [post]
func (h *AdminController) create(c *gin.Context) {
	var in req.CreateAdmin
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	a := m.Admin{FullName: in.FullName, UserID: in.UserID}
	if err := h.DB.Create(&a).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toAdminResp(a))
}

// @Summary List admins
// @Tags admins
// @Produce json
// @Param q query string false "search by name"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.AdminList
// @Router /admins [get]
func (h *AdminController) list(c *gin.Context) {
	var qin req.ListAdminQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Admin{})
	if v := qin.Q; v != "" {
		base = base.Where("full_name ILIKE ?", "%"+v+"%")
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
	var ll []m.Admin
	if err := base.Order("full_name asc").Limit(qin.Limit).Offset(qin.Offset).Find(&ll).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Admin, 0, len(ll))
	for _, a := range ll {
		items = append(items, toAdminResp(a))
	}
	c.JSON(http.StatusOK, resp.List[resp.Admin]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get admin
// @Tags admins
// @Produce json
// @Param id path string true "admin id"
// @Success 200 {object} response.Admin
// @Failure 404 {object} response.Error
// @Router /admins/{id} [get]
func (h *AdminController) get(c *gin.Context) {
	id := c.Param("id")
	var a m.Admin
	if err := h.DB.First(&a, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toAdminResp(a))
}

// @Summary Update admin
// @Tags admins
// @Accept json
// @Produce json
// @Param id path string true "admin id"
// @Param input body request.UpdateAdmin true "admin"
// @Success 200 {object} response.Admin
// @Failure 400 {object} response.Error
// @Router /admins/{id} [put]
func (h *AdminController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateAdmin
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var a m.Admin
	if err := h.DB.First(&a, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.FullName != nil {
		a.FullName = *in.FullName
	}
	if in.UserID != nil {
		a.UserID = in.UserID
	}
	if err := h.DB.Save(&a).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toAdminResp(a))
}

// @Summary Delete admin
// @Tags admins
// @Produce json
// @Param id path string true "admin id"
// @Success 200 {object} map[string]bool
// @Router /admins/{id} [delete]
func (h *AdminController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Admin{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func toAdminResp(a m.Admin) resp.Admin {
	return resp.Admin{ID: a.ID, FullName: a.FullName, UserID: a.UserID}
}
