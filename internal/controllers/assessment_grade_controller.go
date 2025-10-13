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
	"gradeflow/pkg/middleware"
)

type AssessmentGradeController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewAssessmentGradeController(db *gorm.DB, cfg config.Config) *AssessmentGradeController {
	return &AssessmentGradeController{DB: db, Cfg: cfg}
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
	if in.Scale != "" && in.Scale != "points" && in.Scale != "passfail" {
		c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid scale"})
		return
	}
	if in.Scale == "points" && in.ValueNum == nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: "valueNum required for points scale"})
		return
	}
	if in.Scale == "passfail" && in.ValuePass == nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: "valuePass required for passfail scale"})
		return
	}
	var gradedBy string
	if v, ok := c.Get("user"); ok {
		if u, ok2 := v.(*m.User); ok2 {
			gradedBy = u.ID
		}
	}
	d := m.AssessmentGrade{AssessmentID: in.AssessmentID, StudentID: in.StudentID, Scale: in.Scale, ValueNum: in.ValueNum, ValuePass: in.ValuePass, GradedBy: gradedBy, GradedAt: time.Now()}
	if err := h.DB.Create(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toAssessmentGradeResp(d))
}

// @Summary List grades
// @Tags grades
// @Produce json
// @Success 200 {object} response.AssessmentGradeList
// @Router /grades [get]
func (h *AssessmentGradeController) list(c *gin.Context) {
	var qin req.ListAssessmentGradeQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.AssessmentGrade{})
	if v := qin.AssessmentID; v != "" {
		base = base.Where("assessment_id = ?", v)
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
		qin.Limit = 500
	}
	if qin.Limit > 2000 {
		qin.Limit = 2000
	}
	if qin.Offset < 0 {
		qin.Offset = 0
	}
	var dd []m.AssessmentGrade
	if err := base.Limit(qin.Limit).Offset(qin.Offset).Find(&dd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.AssessmentGrade, 0, len(dd))
	for _, d := range dd {
		items = append(items, toAssessmentGradeResp(d))
	}
	c.JSON(http.StatusOK, resp.List[resp.AssessmentGrade]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
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
	var d m.AssessmentGrade
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toAssessmentGradeResp(d))
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
	var d m.AssessmentGrade
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.Scale != "" {
		if in.Scale != "points" && in.Scale != "passfail" {
			c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid scale"})
			return
		}
		d.Scale = in.Scale
	}
	d.ValueNum = in.ValueNum
	d.ValuePass = in.ValuePass
	if err := h.DB.Save(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toAssessmentGradeResp(d))
}

// @Summary Delete grade
// @Tags grades
// @Produce json
// @Param id path string true "grade id"
// @Success 200 {object} map[string]bool
// @Router /grades/{id} [delete]
func (h *AssessmentGradeController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.AssessmentGrade{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
