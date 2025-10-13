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

type StudentController struct {
	DB      *gorm.DB
	Cfg     config.Config
	Service service.StudentService
}

func NewStudentController(db *gorm.DB, cfg config.Config, svc service.StudentService) *StudentController {
	return &StudentController{DB: db, Cfg: cfg, Service: svc}
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
	st, err := h.Service.Create(in)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toStudentResp(*st))
}

// @Summary List students
// @Tags students
// @Produce json
// @Success 200 {object} response.StudentList
// @Router /students [get]
func (h *StudentController) list(c *gin.Context) {
	var qin req.ListStudentQuery
	_ = c.ShouldBindQuery(&qin)
	students, total, appliedLimit, appliedOffset, err := h.Service.List(qin.Limit, qin.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Student, 0, len(students))
	for _, s := range students {
		items = append(items, toStudentResp(s))
	}
    c.JSON(http.StatusOK, resp.List[resp.Student]{Items: items, Page: resp.Page{Limit: appliedLimit, Offset: appliedOffset, Total: total}})
}

// @Summary Get student
// @Tags students
// @Produce json
// @Param id path string true "student id"
// @Success 200 {object} response.Student
// @Router /students/{id} [get]
func (h *StudentController) get(c *gin.Context) {
	id := c.Param("id")
	s, err := h.Service.Get(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, toStudentResp(*s))
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
	s, err := h.Service.Update(id, in)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, toStudentResp(*s))
}

// @Summary Delete student
// @Tags students
// @Produce json
// @Param id path string true "student id"
// @Success 200 {object} map[string]bool
// @Router /students/{id} [delete]
func (h *StudentController) delete(c *gin.Context) {
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

func toStudentResp(s m.Student) resp.Student {
	return resp.Student{
		ID:               s.ID,
		IndividualNumber: s.IndividualNumber,
		FullName:         s.FullName,
		GroupID:          s.GroupID,
		StartYear:        s.StartYear,
		EndYear:          s.EndYear,
	}
}
