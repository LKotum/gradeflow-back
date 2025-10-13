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

type DepartmentController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewDepartmentController(db *gorm.DB, cfg config.Config) *DepartmentController {
	return &DepartmentController{DB: db, Cfg: cfg}
}

func (h *DepartmentController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/departments")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)
}

// create department
// @Summary Create department
// @Tags departments
// @Accept json
// @Produce json
// @Param input body request.CreateDepartment true "department"
// @Success 201 {object} response.Department
// @Failure 400 {object} response.Error
// @Router /departments [post]
func (h *DepartmentController) create(c *gin.Context) {
	var in req.CreateDepartment
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	d := m.Department{Code: in.Code, Name: in.Name}
	if err := h.DB.Create(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toDepartmentResp(d))
}

// list departments
// @Summary List departments
// @Tags departments
// @Produce json
// @Success 200 {object} response.DepartmentList
// @Router /departments [get]
func (h *DepartmentController) list(c *gin.Context) {
	var qin req.ListDepartmentQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Department{})
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
	var dd []m.Department
	if err := base.Order("code asc").Limit(qin.Limit).Offset(qin.Offset).Find(&dd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Department, 0, len(dd))
	for _, d := range dd {
		items = append(items, toDepartmentResp(d))
	}
	c.JSON(http.StatusOK, resp.List[resp.Department]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// get department
// @Summary Get department
// @Tags departments
// @Produce json
// @Param id path string true "department id"
// @Success 200 {object} response.Department
// @Failure 404 {object} response.Error
// @Router /departments/{id} [get]
func (h *DepartmentController) get(c *gin.Context) {
	id := c.Param("id")
	var d m.Department
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toDepartmentResp(d))
}

// update department
// @Summary Update department
// @Tags departments
// @Accept json
// @Produce json
// @Param id path string true "department id"
// @Param input body request.UpdateDepartment true "department"
// @Success 200 {object} response.Department
// @Failure 400 {object} response.Error
// @Router /departments/{id} [put]
func (h *DepartmentController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateDepartment
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var d m.Department
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.Code != "" {
		d.Code = in.Code
	}
	if in.Name != "" {
		d.Name = in.Name
	}
	if err := h.DB.Save(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toDepartmentResp(d))
}

// delete department
// @Summary Delete department
// @Tags departments
// @Produce json
// @Param id path string true "department id"
// @Success 200 {object} response.OK
// @Router /departments/{id} [delete]
func (h *DepartmentController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Department{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.OK{OK: true})
}

func toDepartmentResp(d m.Department) resp.Department {
	return resp.Department{ID: d.ID, Code: d.Code, Name: d.Name}
}
