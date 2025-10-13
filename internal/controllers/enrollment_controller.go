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

type EnrollmentController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewEnrollmentController(db *gorm.DB, cfg config.Config) *EnrollmentController {
	return &EnrollmentController{DB: db, Cfg: cfg}
}

func (h *EnrollmentController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/enrollments")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.StaffOrTeacher(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.StaffOrTeacher(), h.update)
	g.DELETE(":id", middleware.StaffOrTeacher(), h.delete)
}

// create enrollment
// @Summary Create enrollment
// @Tags enrollments
// @Accept json
// @Produce json
// @Param input body request.CreateEnrollment true "enrollment"
// @Success 201 {object} response.Enrollment
// @Failure 400 {object} response.Error
// @Router /enrollments [post]
func (h *EnrollmentController) create(c *gin.Context) {
	var in req.CreateEnrollment
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	d := m.Enrollment{CourseID: in.CourseID, StudentID: in.StudentID}
	if err := h.DB.Create(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toEnrollmentResp(d))
}

// list enrollments
// @Summary List enrollments
// @Tags enrollments
// @Produce json
// @Success 200 {object} response.EnrollmentList
// @Router /enrollments [get]
func (h *EnrollmentController) list(c *gin.Context) {
	var qin req.ListEnrollmentQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Enrollment{})
	if v := qin.CourseID; v != "" {
		base = base.Where("course_id = ?", v)
	}
	if v := qin.StudentID; v != "" {
		base = base.Where("student_id = ?", v)
	}
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
	var dd []m.Enrollment
	if err := base.Limit(qin.Limit).Offset(qin.Offset).Find(&dd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Enrollment, 0, len(dd))
	for _, d := range dd {
		items = append(items, toEnrollmentResp(d))
	}
	c.JSON(http.StatusOK, resp.List[resp.Enrollment]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// get enrollment
// @Summary Get enrollment
// @Tags enrollments
// @Produce json
// @Param id path string true "enrollment id"
// @Success 200 {object} response.Enrollment
// @Failure 404 {object} response.Error
// @Router /enrollments/{id} [get]
func (h *EnrollmentController) get(c *gin.Context) {
	id := c.Param("id")
	var d m.Enrollment
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toEnrollmentResp(d))
}

// update enrollment
// @Summary Update enrollment
// @Tags enrollments
// @Accept json
// @Produce json
// @Param id path string true "enrollment id"
// @Param input body request.UpdateEnrollment true "enrollment"
// @Success 200 {object} response.Enrollment
// @Failure 400 {object} response.Error
// @Router /enrollments/{id} [put]
func (h *EnrollmentController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateEnrollment
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var d m.Enrollment
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.CourseID != "" {
		d.CourseID = in.CourseID
	}
	if in.StudentID != "" {
		d.StudentID = in.StudentID
	}
	if err := h.DB.Save(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toEnrollmentResp(d))
}

// delete enrollment
// @Summary Delete enrollment
// @Tags enrollments
// @Produce json
// @Param id path string true "enrollment id"
// @Success 200 {object} map[string]bool
// @Router /enrollments/{id} [delete]
func (h *EnrollmentController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Enrollment{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func toEnrollmentResp(e m.Enrollment) resp.Enrollment {
	return resp.Enrollment{ID: e.ID, CourseID: e.CourseID, StudentID: e.StudentID}
}
