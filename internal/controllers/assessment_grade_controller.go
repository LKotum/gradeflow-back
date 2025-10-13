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

type AssessmentGradeController struct {
	DB       *gorm.DB
	Cfg      config.Config
	GradeSvc service.AssessmentGradeService
}

func NewAssessmentGradeController(db *gorm.DB, cfg config.Config, gradeSvc service.AssessmentGradeService) *AssessmentGradeController {
	return &AssessmentGradeController{DB: db, Cfg: cfg, GradeSvc: gradeSvc}
}

func (h *AssessmentGradeController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/grades")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.StaffOrTeacher(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.StaffOrTeacher(), h.update)
	g.DELETE(":id", middleware.StaffOrTeacher(), h.delete)
}

// @Summary Create grade
// @Tags grades
// @Accept json
// @Produce json
// @Param input body request.CreateAssessmentGrade true "grade"
// @Success 201 {object} response.AssessmentGrade
// @Failure 400 {object} response.Error
// @Router /grades [post]
func (h *AssessmentGradeController) create(c *gin.Context) {
	var in req.CreateAssessmentGrade
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var gradedBy *string
	if v, ok := c.Get("user"); ok {
		if u, ok2 := v.(*m.User); ok2 {
			gradedBy = new(string)
			*gradedBy = u.ID
		}
	}
	grade, err := h.GradeSvc.Create(in, gradedBy)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, resp.Error{Error: "assessment not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, toAssessmentGradeResp(*grade))
}

// @Summary List grades
// @Tags grades
// @Produce json
// @Param assessmentId query string false "filter by assessment"
// @Param studentId query string false "filter by student"
// @Param courseId query string false "filter by course"
// @Param from query string false "graded from (RFC3339)"
// @Param to query string false "graded to (RFC3339)"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.AssessmentGradeList
// @Router /grades [get]
func (h *AssessmentGradeController) list(c *gin.Context) {
	var qin req.ListAssessmentGradeQuery
	_ = c.ShouldBindQuery(&qin)
	grades, total, limit, offset, err := h.GradeSvc.List(qin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.AssessmentGrade, 0, len(grades))
	for _, g := range grades {
		items = append(items, toAssessmentGradeResp(g))
	}
	c.JSON(http.StatusOK, resp.List[resp.AssessmentGrade]{Items: items, Page: resp.Page{Limit: limit, Offset: offset, Total: total}})
}

// @Summary Get grade
// @Tags grades
// @Produce json
// @Param id path string true "grade id"
// @Success 200 {object} response.AssessmentGrade
// @Failure 404 {object} response.Error
// @Router /grades/{id} [get]
func (h *AssessmentGradeController) get(c *gin.Context) {
	id := c.Param("id")
	grade, err := h.GradeSvc.Get(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, toAssessmentGradeResp(*grade))
}

// @Summary Update grade
// @Tags grades
// @Accept json
// @Produce json
// @Param id path string true "grade id"
// @Param input body request.UpdateAssessmentGrade true "grade"
// @Success 200 {object} response.AssessmentGrade
// @Failure 400 {object} response.Error
// @Router /grades/{id} [put]
func (h *AssessmentGradeController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateAssessmentGrade
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var gradedBy *string
	if v, ok := c.Get("user"); ok {
		if u, ok2 := v.(*m.User); ok2 {
			gradedBy = new(string)
			*gradedBy = u.ID
		}
	}
	grade, err := h.GradeSvc.Update(id, in, gradedBy)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, toAssessmentGradeResp(*grade))
}

// @Summary Delete grade
// @Tags grades
// @Produce json
// @Param id path string true "grade id"
// @Success 200 {object} map[string]bool
// @Router /grades/{id} [delete]
func (h *AssessmentGradeController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.GradeSvc.Delete(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
