package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gradeflow/internal/config"
	req "gradeflow/internal/domain/dto/request"
	resp "gradeflow/internal/domain/dto/response"
	m "gradeflow/internal/domain/models"
	"gradeflow/internal/service"
	"gradeflow/pkg/middleware"
)

type AssessmentController struct {
	DB        *gorm.DB
	Cfg       config.Config
	AssessSvc service.AssessmentService
	GradeSvc  service.AssessmentGradeService
}

func NewAssessmentController(db *gorm.DB, cfg config.Config, assessSvc service.AssessmentService, gradeSvc service.AssessmentGradeService) *AssessmentController {
	return &AssessmentController{DB: db, Cfg: cfg, AssessSvc: assessSvc, GradeSvc: gradeSvc}
}

func (h *AssessmentController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/assessments")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.StaffOrTeacher(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.StaffOrTeacher(), h.update)
	g.DELETE(":id", middleware.StaffOrTeacher(), h.delete)

	g.GET(":id/grades", h.listGrades)
	g.POST(":id/grades/bulk", middleware.StaffOrTeacher(), h.bulkGrades)
}

// @Summary Create assessment
// @Tags assessments
// @Accept json
// @Produce json
// @Param input body request.CreateAssessment true "assessment"
// @Success 201 {object} response.Assessment
// @Failure 400 {object} response.Error
// @Router /assessments [post]
func (h *AssessmentController) create(c *gin.Context) {
	var in req.CreateAssessment
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	assessment, err := h.AssessSvc.Create(in)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toAssessmentResp(*assessment))
}

// @Summary List assessments
// @Tags assessments
// @Produce json
// @Param courseId query string false "filter by course"
// @Param from query string false "date from (RFC3339)"
// @Param to query string false "date to (RFC3339)"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.AssessmentList
// @Router /assessments [get]
func (h *AssessmentController) list(c *gin.Context) {
	var qin req.ListAssessmentQuery
	_ = c.ShouldBindQuery(&qin)
	assessments, total, limit, offset, err := h.AssessSvc.List(qin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Assessment, 0, len(assessments))
	for _, a := range assessments {
		items = append(items, toAssessmentResp(a))
	}
	c.JSON(http.StatusOK, resp.List[resp.Assessment]{Items: items, Page: resp.Page{Limit: limit, Offset: offset, Total: total}})
}

// @Summary Get assessment
// @Tags assessments
// @Produce json
// @Param id path string true "assessment id"
// @Success 200 {object} response.Assessment
// @Failure 404 {object} response.Error
// @Router /assessments/{id} [get]
func (h *AssessmentController) get(c *gin.Context) {
	id := c.Param("id")
	assessment, err := h.AssessSvc.Get(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, toAssessmentResp(*assessment))
}

// @Summary Update assessment
// @Tags assessments
// @Accept json
// @Produce json
// @Param id path string true "assessment id"
// @Param input body request.UpdateAssessment true "assessment"
// @Success 200 {object} response.Assessment
// @Failure 400 {object} response.Error
// @Router /assessments/{id} [put]
func (h *AssessmentController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateAssessment
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	assessment, err := h.AssessSvc.Update(id, in)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, toAssessmentResp(*assessment))
}

// @Summary Delete assessment
// @Tags assessments
// @Produce json
// @Param id path string true "assessment id"
// @Success 200 {object} map[string]bool
// @Router /assessments/{id} [delete]
func (h *AssessmentController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.AssessSvc.Delete(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// @Summary List grades for assessment
// @Tags assessments,grades
// @Produce json
// @Param id path string true "assessment id"
// @Success 200 {object} response.AssessmentGradeList
// @Router /assessments/{id}/grades [get]
func (h *AssessmentController) listGrades(c *gin.Context) {
	aid := c.Param("id")
	var q req.PaginationQuery
	_ = c.ShouldBindQuery(&q)
	grades, total, limit, offset, err := h.GradeSvc.ListByAssessment(aid, q.Limit, q.Offset)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "assessment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		}
		return
	}
	items := make([]resp.AssessmentGrade, 0, len(grades))
	for _, g := range grades {
		items = append(items, toAssessmentGradeResp(g))
	}
	c.JSON(http.StatusOK, resp.List[resp.AssessmentGrade]{Items: items, Page: resp.Page{Limit: limit, Offset: offset, Total: total}})
}

// @Summary Bulk upsert grades for assessment
// @Tags assessments,grades
// @Accept json
// @Produce json
// @Param id path string true "assessment id"
// @Param input body []request.GradeItem true "grades"
// @Success 200 {object} response.AssessmentGradeList
// @Failure 400 {object} response.Error
// @Router /assessments/{id}/grades/bulk [post]
func (h *AssessmentController) bulkGrades(c *gin.Context) {
	aid := c.Param("id")
	var in []req.GradeItem
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
	grades, err := h.GradeSvc.BulkUpsert(aid, in, gradedBy)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		} else {
			c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		}
		return
	}
	items := make([]resp.AssessmentGrade, 0, len(grades))
	for _, g := range grades {
		items = append(items, toAssessmentGradeResp(g))
	}
	c.JSON(http.StatusOK, resp.List[resp.AssessmentGrade]{Items: items, Page: resp.Page{Limit: len(items), Offset: 0, Total: int64(len(items))}})
}

func toAssessmentResp(a m.Assessment) resp.Assessment {
	return resp.Assessment{ID: a.ID, CourseID: a.CourseID, Type: a.Type, DateAt: a.DateAt, Room: a.Room, Scale: a.Scale, MaxPts: a.MaxPts}
}

func toAssessmentGradeResp(g m.AssessmentGrade) resp.AssessmentGrade {
	return resp.AssessmentGrade{ID: g.ID, AssessmentID: g.AssessmentID, StudentID: g.StudentID, Scale: g.Scale, ValueNum: g.ValueNum, ValuePass: g.ValuePass, GradedBy: g.GradedBy, GradedAt: g.GradedAt.Format(time.RFC3339)}
}
