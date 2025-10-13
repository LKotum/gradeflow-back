package controllers

import (
	"errors"
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

type ExamAttemptController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewExamAttemptController(db *gorm.DB, cfg config.Config) *ExamAttemptController {
	return &ExamAttemptController{DB: db, Cfg: cfg}
}

func (h *ExamAttemptController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/exam-attempts")
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)
}

// @Summary Create exam attempt
// @Tags exam-attempts
// @Accept json
// @Produce json
// @Param input body request.CreateExamAttempt true "attempt"
// @Success 201 {object} response.ExamAttempt
// @Failure 400 {object} response.Error
// @Router /exam-attempts [post]
func (h *ExamAttemptController) create(c *gin.Context) {
	var in req.CreateExamAttempt
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	if err := validateScale(in.ResultScale, in.ValueNum, in.ValuePass); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	ea := m.ExamAttempt{AssessmentID: in.AssessmentID, StudentID: in.StudentID, AttemptNo: in.AttemptNo, ResultScale: in.ResultScale, ValueNum: in.ValueNum, ValuePass: in.ValuePass, Notes: in.Notes}
	if in.DateAt != nil {
		ea.DateAt = parseTimePtr(*in.DateAt)
	}
	if err := h.DB.Create(&ea).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toExamAttemptResp(ea))
}

// @Summary List exam attempts
// @Tags exam-attempts
// @Produce json
// @Param assessmentId query string false "filter by assessment"
// @Param studentId query string false "filter by student"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.ExamAttemptList
// @Router /exam-attempts [get]
func (h *ExamAttemptController) list(c *gin.Context) {
	var qin req.ListExamAttemptQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.ExamAttempt{})
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
		qin.Limit = 50
	}
	if qin.Limit > 200 {
		qin.Limit = 200
	}
	if qin.Offset < 0 {
		qin.Offset = 0
	}
	var ll []m.ExamAttempt
	if err := base.Order("date_at asc nulls last").Limit(qin.Limit).Offset(qin.Offset).Find(&ll).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.ExamAttempt, 0, len(ll))
	for _, e := range ll {
		items = append(items, toExamAttemptResp(e))
	}
	c.JSON(http.StatusOK, resp.List[resp.ExamAttempt]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get exam attempt
// @Tags exam-attempts
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} response.ExamAttempt
// @Failure 404 {object} response.Error
// @Router /exam-attempts/{id} [get]
func (h *ExamAttemptController) get(c *gin.Context) {
	id := c.Param("id")
	var ea m.ExamAttempt
	if err := h.DB.First(&ea, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toExamAttemptResp(ea))
}

// @Summary Update exam attempt
// @Tags exam-attempts
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param input body request.UpdateExamAttempt true "attempt"
// @Success 200 {object} response.ExamAttempt
// @Failure 400 {object} response.Error
// @Router /exam-attempts/{id} [put]
func (h *ExamAttemptController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateExamAttempt
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var ea m.ExamAttempt
	if err := h.DB.First(&ea, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.AttemptNo != nil {
		ea.AttemptNo = *in.AttemptNo
	}
	if in.DateAt != nil {
		ea.DateAt = parseTimePtr(*in.DateAt)
	}
	if in.ResultScale != nil {
		if err := validateScale(*in.ResultScale, in.ValueNum, in.ValuePass); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ea.ResultScale = *in.ResultScale
	}
	if in.ValueNum != nil || in.ValuePass != nil {
		rs := ea.ResultScale
		if in.ResultScale != nil {
			rs = *in.ResultScale
		}
		if err := validateScale(rs, in.ValueNum, in.ValuePass); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ea.ValueNum = in.ValueNum
		ea.ValuePass = in.ValuePass
	}
	if in.Notes != nil {
		ea.Notes = *in.Notes
	}
	if err := h.DB.Save(&ea).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toExamAttemptResp(ea))
}

// @Summary Delete exam attempt
// @Tags exam-attempts
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} map[string]bool
// @Router /exam-attempts/{id} [delete]
func (h *ExamAttemptController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.ExamAttempt{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func validateScale(scale string, vnum *int, vpass *bool) error {
	switch scale {
	case "passfail":
		if vpass == nil || vnum != nil {
			return errors.New("invalid value for passfail")
		}
	case "five":
		if vnum == nil || vpass != nil {
			return errors.New("invalid value for five")
		}
		if *vnum < 2 || *vnum > 5 {
			return errors.New("value must be 2..5")
		}
	case "hundred":
		if vnum == nil || vpass != nil {
			return errors.New("invalid value for hundred")
		}
		if *vnum < 0 || *vnum > 100 {
			return errors.New("value must be 0..100")
		}
	default:
		return errors.New("invalid scale")
	}
	return nil
}

func toExamAttemptResp(e m.ExamAttempt) resp.ExamAttempt {
	var date *string
	if e.DateAt != nil {
		s := e.DateAt.Format(time.RFC3339)
		date = &s
	}
	return resp.ExamAttempt{ID: e.ID, AssessmentID: e.AssessmentID, StudentID: e.StudentID, AttemptNo: e.AttemptNo, DateAt: date, ResultScale: e.ResultScale, ValueNum: e.ValueNum, ValuePass: e.ValuePass, Notes: e.Notes}
}
