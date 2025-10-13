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

type StaffController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewStaffController(db *gorm.DB, cfg config.Config) *StaffController {
	return &StaffController{DB: db, Cfg: cfg}
}

func (h *StaffController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/staff")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)
}

// @Summary Create staff member
// @Tags staff
// @Accept json
// @Produce json
// @Param input body request.CreateStaff true "staff"
// @Success 201 {object} response.Staff
// @Failure 400 {object} response.Error
// @Router /staff [post]
func (h *StaffController) create(c *gin.Context) {
	var in req.CreateStaff
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	s := m.StaffMember{FullName: in.FullName, DepartmentID: in.DepartmentID, Position: in.Position, UserID: in.UserID}
	if err := h.DB.Create(&s).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toStaffResp(s))
}

// @Summary List staff
// @Tags staff
// @Produce json
// @Param departmentId query string false "filter by department"
// @Param q query string false "search by name"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.StaffList
// @Router /staff [get]
func (h *StaffController) list(c *gin.Context) {
	var qin req.ListStaffQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.StaffMember{})
	if v := qin.DepartmentID; v != "" {
		base = base.Where("department_id = ?", v)
	}
	if v := qin.Q; v != "" {
		base = base.Where("full_name ILIKE ?", "%"+v+"%")
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	// pagination defaults
	if qin.Limit <= 0 {
		qin.Limit = 50
	}
	if qin.Limit > 200 {
		qin.Limit = 200
	}
	if qin.Offset < 0 {
		qin.Offset = 0
	}
	var ll []m.StaffMember
	if err := base.Order("full_name asc").Limit(qin.Limit).Offset(qin.Offset).Find(&ll).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Staff, 0, len(ll))
	for _, s := range ll {
		items = append(items, toStaffResp(s))
	}
	c.JSON(http.StatusOK, resp.List[resp.Staff]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get staff member
// @Tags staff
// @Produce json
// @Param id path string true "staff id"
// @Success 200 {object} response.Staff
// @Failure 404 {object} response.Error
// @Router /staff/{id} [get]
func (h *StaffController) get(c *gin.Context) {
	id := c.Param("id")
	var s m.StaffMember
	if err := h.DB.First(&s, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toStaffResp(s))
}

// @Summary Update staff member
// @Tags staff
// @Accept json
// @Produce json
// @Param id path string true "staff id"
// @Param input body request.UpdateStaff true "staff"
// @Success 200 {object} response.Staff
// @Failure 400 {object} response.Error
// @Router /staff/{id} [put]
func (h *StaffController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateStaff
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var s m.StaffMember
	if err := h.DB.First(&s, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.FullName != nil {
		s.FullName = *in.FullName
	}
	if in.DepartmentID != nil {
		s.DepartmentID = in.DepartmentID
	}
	if in.Position != nil {
		s.Position = *in.Position
	}
	if in.UserID != nil {
		s.UserID = in.UserID
	}
	if err := h.DB.Save(&s).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toStaffResp(s))
}

// @Summary Delete staff member
// @Tags staff
// @Produce json
// @Param id path string true "staff id"
// @Success 200 {object} map[string]bool
// @Router /staff/{id} [delete]
func (h *StaffController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.StaffMember{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func toStaffResp(s m.StaffMember) resp.Staff {
	return resp.Staff{ID: s.ID, FullName: s.FullName, DepartmentID: s.DepartmentID, Position: s.Position, UserID: s.UserID}
}
