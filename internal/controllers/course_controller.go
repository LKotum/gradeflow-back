package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gradeflow/internal/config"
	req "gradeflow/internal/domain/dto/request"
	resp "gradeflow/internal/domain/dto/response"
	m "gradeflow/internal/domain/models"
	"gradeflow/internal/service"
	"gradeflow/pkg/middleware"
)

type CourseController struct {
	DB      *gorm.DB
	Cfg     config.Config
	Service service.CourseService
}

func NewCourseController(db *gorm.DB, cfg config.Config, svc service.CourseService) *CourseController {
    return &CourseController{DB: db, Cfg: cfg, Service: svc}
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
	course, err := h.Service.Create(in)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toCourseResp(*course))
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
	courses, total, limit, offset, err := h.Service.List(qin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Course, 0, len(courses))
	for _, d := range courses {
		items = append(items, toCourseResp(d))
	}
	c.JSON(http.StatusOK, resp.List[resp.Course]{Items: items, Page: resp.Page{Limit: limit, Offset: offset, Total: total}})
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
	course, err := h.Service.Get(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, toCourseResp(*course))
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
	course, err := h.Service.Update(id, in)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, toCourseResp(*course))
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
	if err := h.Service.Delete(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
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
