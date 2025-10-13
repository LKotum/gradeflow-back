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
	"gradeflow/pkg/utils"
)

type TeacherController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewTeacherController(db *gorm.DB, cfg config.Config) *TeacherController {
	return &TeacherController{DB: db, Cfg: cfg}
}

func (h *TeacherController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/teachers")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)
}

// @Summary Create teacher
// @Tags teachers
// @Accept json
// @Produce json
// @Param input body request.CreateTeacher true "teacher"
// @Success 201 {object} response.Teacher
// @Failure 400 {object} response.Error
// @Router /teachers [post]
func (h *TeacherController) create(c *gin.Context) {
	var in req.CreateTeacher
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	t := m.Teacher{FullName: in.FullName, DepartmentID: in.DepartmentID, Title: in.Title, Rank: in.Rank, UserID: in.UserID}
	if err := h.DB.Create(&t).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toTeacherResp(t))
}

// @Summary List teachers
// @Tags teachers
// @Produce json
// @Param departmentId query string false "filter by department"
// @Param q query string false "search by name"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.TeacherList
// @Router /teachers [get]
func (h *TeacherController) list(c *gin.Context) {
	var qin req.ListTeacherQuery
	_ = c.ShouldBindQuery(&qin)

	base := h.DB.Model(&m.Teacher{})
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
	var ll []m.Teacher
	q := utils.ApplyPagination(base.Order("full_name asc"), c, 200, 50)
	if err := q.Find(&ll).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Teacher, 0, len(ll))
	for _, t := range ll {
		items = append(items, toTeacherResp(t))
	}
	c.JSON(http.StatusOK, resp.List[resp.Teacher]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get teacher
// @Tags teachers
// @Produce json
// @Param id path string true "teacher id"
// @Success 200 {object} response.Teacher
// @Failure 404 {object} response.Error
// @Router /teachers/{id} [get]
func (h *TeacherController) get(c *gin.Context) {
	id := c.Param("id")
	var t m.Teacher
	if err := h.DB.First(&t, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toTeacherResp(t))
}

// @Summary Update teacher
// @Tags teachers
// @Accept json
// @Produce json
// @Param id path string true "teacher id"
// @Param input body request.UpdateTeacher true "teacher"
// @Success 200 {object} response.Teacher
// @Failure 400 {object} response.Error
// @Router /teachers/{id} [put]
func (h *TeacherController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateTeacher
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var t m.Teacher
	if err := h.DB.First(&t, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.FullName != nil {
		t.FullName = *in.FullName
	}
	if in.DepartmentID != nil {
		t.DepartmentID = in.DepartmentID
	}
	if in.Title != nil {
		t.Title = *in.Title
	}
	if in.Rank != nil {
		t.Rank = *in.Rank
	}
	if in.UserID != nil {
		t.UserID = in.UserID
	}
	if err := h.DB.Save(&t).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toTeacherResp(t))
}

// @Summary Delete teacher
// @Tags teachers
// @Produce json
// @Param id path string true "teacher id"
// @Success 200 {object} map[string]bool
// @Router /teachers/{id} [delete]
func (h *TeacherController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Teacher{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func toTeacherResp(t m.Teacher) resp.Teacher {
	return resp.Teacher{ID: t.ID, FullName: t.FullName, DepartmentID: t.DepartmentID, Title: t.Title, Rank: t.Rank, UserID: t.UserID}
}
