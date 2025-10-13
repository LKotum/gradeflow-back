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

type EnrollmentController struct {
	DB      *gorm.DB
	Cfg     config.Config
	Service service.EnrollmentService
}

func NewEnrollmentController(db *gorm.DB, cfg config.Config, svc service.EnrollmentService) *EnrollmentController {
	return &EnrollmentController{DB: db, Cfg: cfg, Service: svc}
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
	enrollment, err := h.Service.Create(in)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toEnrollmentResp(*enrollment))
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
	enrollments, total, limit, offset, err := h.Service.List(qin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Enrollment, 0, len(enrollments))
	for _, d := range enrollments {
		items = append(items, toEnrollmentResp(d))
	}
	c.JSON(http.StatusOK, resp.List[resp.Enrollment]{Items: items, Page: resp.Page{Limit: limit, Offset: offset, Total: total}})
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
	enrollment, err := h.Service.Get(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, toEnrollmentResp(*enrollment))
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
	enrollment, err := h.Service.Update(id, in)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, toEnrollmentResp(*enrollment))
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

func toEnrollmentResp(e m.Enrollment) resp.Enrollment {
	return resp.Enrollment{ID: e.ID, CourseID: e.CourseID, StudentID: e.StudentID}
}
