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

type CourseController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewCourseController(db *gorm.DB, cfg config.Config) *CourseController {
	return &CourseController{DB: db, Cfg: cfg}
}

func (h *CourseController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/courses")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)
}

// create course
// @Summary Create course
// @Tags courses
// @Accept json
// @Produce json
// @Param input body request.CreateCourse true "course"
// @Success 201 {object} response.Course
// @Failure 400 {object} response.Error
// @Router /courses [post]
func (h *CourseController) create(c *gin.Context) {
	var in req.CreateCourse
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	d := m.Course{
		SubjectID:         in.SubjectID,
		DepartmentID:      in.DepartmentID,
		ProgramID:         in.ProgramID,
		AcademicSessionID: in.AcademicSessionID,
		Title:             in.Title,
		TeacherID:         in.TeacherID,
		Room:              in.Room,
	}
	if err := h.DB.Create(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toCourseResp(d))
}

// list courses
// @Summary List courses
// @Tags courses
// @Produce json
// @Success 200 {object} response.CourseList
// @Router /courses [get]
func (h *CourseController) list(c *gin.Context) {
	var qin req.ListCourseQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Course{})
	if v := qin.DepartmentID; v != "" {
		base = base.Where("department_id = ?", v)
	}
	if v := qin.ProgramID; v != "" {
		base = base.Where("program_id = ?", v)
	}
	if v := qin.SubjectID; v != "" {
		base = base.Where("subject_id = ?", v)
	}
	if v := qin.AcademicSessionID; v != "" {
		base = base.Where("academic_session_id = ?", v)
	}
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
	var dd []m.Course
	if err := base.Order("title asc").Limit(qin.Limit).Offset(qin.Offset).Find(&dd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Course, 0, len(dd))
	for _, d := range dd {
		items = append(items, toCourseResp(d))
	}
	c.JSON(http.StatusOK, resp.List[resp.Course]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// get course
// @Summary Get course
// @Tags courses
// @Produce json
// @Param id path string true "course id"
// @Success 200 {object} response.Course
// @Failure 404 {object} response.Error
// @Router /courses/{id} [get]
func (h *CourseController) get(c *gin.Context) {
	id := c.Param("id")
	var d m.Course
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toCourseResp(d))
}

// update course
// @Summary Update course
// @Tags courses
// @Accept json
// @Produce json
// @Param id path string true "course id"
// @Param input body request.UpdateCourse true "course"
// @Success 200 {object} response.Course
// @Failure 400 {object} response.Error
// @Router /courses/{id} [put]
func (h *CourseController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateCourse
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var d m.Course
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.SubjectID != "" {
		d.SubjectID = in.SubjectID
	}
	if in.DepartmentID != "" {
		d.DepartmentID = in.DepartmentID
	}
	if in.ProgramID != nil {
		d.ProgramID = in.ProgramID
	}
	if in.AcademicSessionID != "" {
		d.AcademicSessionID = in.AcademicSessionID
	}
	if in.Title != "" {
		d.Title = in.Title
	}
	if in.TeacherID != nil {
		d.TeacherID = in.TeacherID
	}
	if in.Room != "" {
		d.Room = in.Room
	}
	if err := h.DB.Save(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toCourseResp(d))
}

// delete course
// @Summary Delete course
// @Tags courses
// @Produce json
// @Param id path string true "course id"
// @Success 200 {object} map[string]bool
// @Router /courses/{id} [delete]
func (h *CourseController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Course{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func toCourseResp(c0 m.Course) resp.Course {
	return resp.Course{
		ID:                c0.ID,
		SubjectID:         c0.SubjectID,
		DepartmentID:      c0.DepartmentID,
		ProgramID:         c0.ProgramID,
		AcademicSessionID: c0.AcademicSessionID,
		Title:             c0.Title,
		TeacherID:         c0.TeacherID,
		Room:              c0.Room,
	}
}
