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

type StudentController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewStudentController(db *gorm.DB, cfg config.Config) *StudentController {
	return &StudentController{DB: db, Cfg: cfg}
}

func (h *StudentController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/students")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)
}

// @Summary Create student
// @Tags students
// @Accept json
// @Produce json
// @Param input body request.CreateStudent true "student"
// @Success 201 {object} response.Student
// @Router /students [post]
func (h *StudentController) create(c *gin.Context) {
	var in req.CreateStudent
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var gid *string
	if in.GroupID != "" {
		gid = &in.GroupID
	}
	s := m.Student{IndividualNumber: in.IndividualNumber, FullName: in.FullName, GroupID: gid}
	if err := h.DB.Create(&s).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toStudentResp(s))
}

// @Summary List students
// @Tags students
// @Produce json
// @Success 200 {object} response.StudentList
// @Router /students [get]
func (h *StudentController) list(c *gin.Context) {
	var qin req.ListStudentQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Student{})
	var total int64
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if qin.Limit <= 0 {
		qin.Limit = 200
	}
	if qin.Limit > 1000 {
		qin.Limit = 1000
	}
	if qin.Offset < 0 {
		qin.Offset = 0
	}
	var ss []m.Student
	if err := base.Order("full_name asc").Limit(qin.Limit).Offset(qin.Offset).Find(&ss).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Student, 0, len(ss))
	for _, s := range ss {
		items = append(items, toStudentResp(s))
	}
	c.JSON(http.StatusOK, resp.List[resp.Student]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get student
// @Tags students
// @Produce json
// @Param id path string true "student id"
// @Success 200 {object} response.Student
// @Router /students/{id} [get]
func (h *StudentController) get(c *gin.Context) {
	id := c.Param("id")
	var s m.Student
	if err := h.DB.First(&s, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toStudentResp(s))
}

// @Summary Update student
// @Tags students
// @Accept json
// @Produce json
// @Param id path string true "student id"
// @Param input body request.UpdateStudent true "student"
// @Success 200 {object} response.Student
// @Router /students/{id} [put]
func (h *StudentController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateStudent
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var s m.Student
	if err := h.DB.First(&s, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.IndividualNumber != "" {
		s.IndividualNumber = in.IndividualNumber
	}
	if in.FullName != "" {
		s.FullName = in.FullName
	}
	if in.GroupID != "" {
		s.GroupID = &in.GroupID
	}
	if err := h.DB.Save(&s).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toStudentResp(s))
}

// @Summary Delete student
// @Tags students
// @Produce json
// @Param id path string true "student id"
// @Success 200 {object} map[string]bool
// @Router /students/{id} [delete]
func (h *StudentController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Student{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func toStudentResp(s m.Student) resp.Student {
	return resp.Student{ID: s.ID, IndividualNumber: s.IndividualNumber, FullName: s.FullName, GroupID: s.GroupID}
}
