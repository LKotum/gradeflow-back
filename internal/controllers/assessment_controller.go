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

type AssessmentController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewAssessmentController(db *gorm.DB, cfg config.Config) *AssessmentController {
	return &AssessmentController{DB: db, Cfg: cfg}
}

func (h *AssessmentController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/assessments")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.StaffOrTeacher(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.StaffOrTeacher(), h.update)
	g.DELETE(":id", middleware.StaffOrTeacher(), h.delete)

	// nested: manage grades
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
	if in.Scale != "points" && in.Scale != "passfail" {
		c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid scale"})
		return
	}
	d := m.Assessment{CourseID: in.CourseID, Type: in.Type, DateAt: in.DateAt, Room: in.Room, Scale: in.Scale, MaxPts: in.MaxPts}
	if err := h.DB.Create(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toAssessmentResp(d))
}

// @Summary List assessments
// @Tags assessments
// @Produce json
// @Success 200 {object} response.AssessmentList
// @Router /assessments [get]
func (h *AssessmentController) list(c *gin.Context) {
	var qin req.ListAssessmentQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Assessment{})
	if v := qin.CourseID; v != "" {
		base = base.Where("course_id = ?", v)
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
	var dd []m.Assessment
	if err := base.Order("date_at asc").Limit(qin.Limit).Offset(qin.Offset).Find(&dd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Assessment, 0, len(dd))
	for _, d := range dd {
		items = append(items, toAssessmentResp(d))
	}
	c.JSON(http.StatusOK, resp.List[resp.Assessment]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
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
	var d m.Assessment
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toAssessmentResp(d))
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
	var d m.Assessment
	if err := h.DB.First(&d, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.CourseID != "" {
		d.CourseID = in.CourseID
	}
	if in.Type != "" {
		d.Type = in.Type
	}
	if in.DateAt != nil {
		d.DateAt = *in.DateAt
	}
	if in.Room != "" {
		d.Room = in.Room
	}
	if in.Scale != "" {
		if in.Scale != "points" && in.Scale != "passfail" {
			c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid scale"})
			return
		}
		d.Scale = in.Scale
	}
	if in.MaxPts != nil {
		d.MaxPts = in.MaxPts
	}
	if err := h.DB.Save(&d).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toAssessmentResp(d))
}

// @Summary Delete assessment
// @Tags assessments
// @Produce json
// @Param id path string true "assessment id"
// @Success 200 {object} map[string]bool
// @Router /assessments/{id} [delete]
func (h *AssessmentController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Assessment{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
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
	base := h.DB.Model(&m.AssessmentGrade{}).Where("assessment_id = ?", aid)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if q.Limit <= 0 {
		q.Limit = 500
	}
	if q.Limit > 2000 {
		q.Limit = 2000
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	var gg []m.AssessmentGrade
	if err := base.Limit(q.Limit).Offset(q.Offset).Find(&gg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.AssessmentGrade, 0, len(gg))
	for _, g := range gg {
		items = append(items, toAssessmentGradeResp(g))
	}
	c.JSON(http.StatusOK, resp.List[resp.AssessmentGrade]{Items: items, Page: resp.Page{Limit: q.Limit, Offset: q.Offset, Total: total}})
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
	var gradedBy string
	if v, ok := c.Get("user"); ok {
		if u, ok2 := v.(*m.User); ok2 {
			gradedBy = u.ID
		}
	}
	now := time.Now()
	for _, e := range in {
		if e.Scale != "" && e.Scale != "points" && e.Scale != "passfail" {
			c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid scale"})
			return
		}
		if e.Scale == "points" && e.ValueNum == nil {
			c.JSON(http.StatusBadRequest, resp.Error{Error: "valueNum required for points scale"})
			return
		}
		if e.Scale == "passfail" && e.ValuePass == nil {
			c.JSON(http.StatusBadRequest, resp.Error{Error: "valuePass required for passfail scale"})
			return
		}
		var rec m.AssessmentGrade
		if err := h.DB.Where("assessment_id = ? AND student_id = ?", aid, e.StudentID).First(&rec).Error; err == nil {
			if e.Scale != "" {
				rec.Scale = e.Scale
			}
			rec.ValueNum = e.ValueNum
			rec.ValuePass = e.ValuePass
			if gradedBy != "" {
				rec.GradedBy = gradedBy
			}
			rec.GradedAt = now
			_ = h.DB.Save(&rec).Error
		} else {
			rec = m.AssessmentGrade{AssessmentID: aid, StudentID: e.StudentID, Scale: e.Scale, ValueNum: e.ValueNum, ValuePass: e.ValuePass, GradedBy: gradedBy, GradedAt: now}
			_ = h.DB.Create(&rec).Error
		}
	}
	var out []m.AssessmentGrade
	_ = h.DB.Where("assessment_id = ?", aid).Find(&out).Error
	items := make([]resp.AssessmentGrade, 0, len(out))
	for _, g := range out {
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
