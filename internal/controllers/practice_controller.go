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

type PracticeController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewPracticeController(db *gorm.DB, cfg config.Config) *PracticeController {
	return &PracticeController{DB: db, Cfg: cfg}
}

func (h *PracticeController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/practices")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)

	g.GET(":id/enrollments", h.listEnrollments)
	g.POST(":id/enrollments/bulk", middleware.AdminOrDean(), h.bulkUpsertEnrollments)
}

// @Summary Create practice
// @Tags practices
// @Accept json
// @Produce json
// @Param input body request.CreatePractice true "practice"
// @Success 201 {object} response.Practice
// @Failure 400 {object} response.Error
// @Router /practices [post]
func (h *PracticeController) create(c *gin.Context) {
	var in req.CreatePractice
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	p := m.Practice{Title: in.Title, Description: in.Description, DepartmentID: in.DepartmentID, ProgramID: in.ProgramID, StartDate: in.StartDate, EndDate: in.EndDate, SupervisorID: in.SupervisorID}
	if err := h.DB.Create(&p).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toPracticeResp(p))
}

// @Summary List practices
// @Tags practices
// @Produce json
// @Param departmentId query string false "filter by department"
// @Param programId query string false "filter by program"
// @Param from query string false "start date >= (RFC3339)"
// @Param to query string false "end date <= (RFC3339)"
// @Param q query string false "search by title"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.PracticeList
// @Router /practices [get]
func (h *PracticeController) list(c *gin.Context) {
	var qin req.ListPracticeQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Practice{})
	if v := qin.DepartmentID; v != "" {
		base = base.Where("department_id = ?", v)
	}
	if v := qin.ProgramID; v != "" {
		base = base.Where("program_id = ?", v)
	}
	if v := qin.Q; v != "" {
		base = base.Where("title ILIKE ?", "%"+v+"%")
	}
	if v := qin.From; v != "" {
		if tm, err := time.Parse(time.RFC3339, v); err == nil {
			base = base.Where("start_date >= ?", tm)
		}
	}
	if v := qin.To; v != "" {
		if tm, err := time.Parse(time.RFC3339, v); err == nil {
			base = base.Where("end_date <= ?", tm)
		}
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
	var ll []m.Practice
	if err := base.Order("start_date asc nulls last").Limit(qin.Limit).Offset(qin.Offset).Find(&ll).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Practice, 0, len(ll))
	for _, p := range ll {
		items = append(items, toPracticeResp(p))
	}
	c.JSON(http.StatusOK, resp.List[resp.Practice]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get practice
// @Tags practices
// @Produce json
// @Param id path string true "practice id"
// @Success 200 {object} response.Practice
// @Failure 404 {object} response.Error
// @Router /practices/{id} [get]
func (h *PracticeController) get(c *gin.Context) {
	id := c.Param("id")
	var p m.Practice
	if err := h.DB.First(&p, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toPracticeResp(p))
}

// @Summary Update practice
// @Tags practices
// @Accept json
// @Produce json
// @Param id path string true "practice id"
// @Param input body request.UpdatePractice true "practice"
// @Success 200 {object} response.Practice
// @Failure 400 {object} response.Error
// @Router /practices/{id} [put]
func (h *PracticeController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdatePractice
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var p m.Practice
	if err := h.DB.First(&p, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.Title != nil {
		p.Title = *in.Title
	}
	if in.Description != nil {
		p.Description = *in.Description
	}
	if in.DepartmentID != nil {
		p.DepartmentID = in.DepartmentID
	}
	if in.ProgramID != nil {
		p.ProgramID = in.ProgramID
	}
	if in.StartDate != nil {
		p.StartDate = in.StartDate
	}
	if in.EndDate != nil {
		p.EndDate = in.EndDate
	}
	if in.SupervisorID != nil {
		p.SupervisorID = in.SupervisorID
	}
	if err := h.DB.Save(&p).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toPracticeResp(p))
}

// @Summary Delete practice
// @Tags practices
// @Produce json
// @Param id path string true "practice id"
// @Success 200 {object} map[string]bool
// @Router /practices/{id} [delete]
func (h *PracticeController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Practice{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// @Summary List practice enrollments
// @Tags practices
// @Produce json
// @Param id path string true "practice id"
// @Param studentId query string false "filter by student"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.PracticeEnrollmentList
// @Router /practices/{id}/enrollments [get]
func (h *PracticeController) listEnrollments(c *gin.Context) {
	pid := c.Param("id")
	var qin struct {
		req.PaginationQuery
		StudentID string `form:"studentId"`
	}
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.PracticeEnrollment{}).Where("practice_id = ?", pid)
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
	var ll []m.PracticeEnrollment
	if err := base.Order("created_at asc").Limit(qin.Limit).Offset(qin.Offset).Find(&ll).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.PracticeEnrollment, 0, len(ll))
	for _, e := range ll {
		items = append(items, toPracticeEnrollmentResp(e))
	}
	c.JSON(http.StatusOK, resp.List[resp.PracticeEnrollment]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Bulk upsert enrollments for practice
// @Tags practices
// @Accept json
// @Produce json
// @Param id path string true "practice id"
// @Param input body []request.PracticeEnrollmentItem true "enrollments"
// @Success 200 {array} response.PracticeEnrollment
// @Failure 400 {object} response.Error
// @Router /practices/{id}/enrollments/bulk [post]
func (h *PracticeController) bulkUpsertEnrollments(c *gin.Context) {
	pid := c.Param("id")
	var in []req.PracticeEnrollmentItem
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	// simple validation
	for _, e := range in {
		if e.Status != "" && e.Status != "enrolled" && e.Status != "completed" && e.Status != "failed" {
			c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid status"})
			return
		}
	}
	res := make([]m.PracticeEnrollment, 0, len(in))
	tx := h.DB.Begin()
	for _, e := range in {
		var pe m.PracticeEnrollment
		if err := tx.Where("practice_id = ? and student_id = ?", pid, e.StudentID).First(&pe).Error; err != nil {
			// create
			pe = m.PracticeEnrollment{PracticeID: pid, StudentID: e.StudentID, Place: e.Place}
			if e.Status != "" {
				pe.Status = e.Status
			}
			if err := tx.Create(&pe).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
				return
			}
		} else {
			// update
			if e.Place != "" {
				pe.Place = e.Place
			}
			if e.Status != "" {
				pe.Status = e.Status
			}
			if err := tx.Save(&pe).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
				return
			}
		}
		res = append(res, pe)
	}
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	out := make([]resp.PracticeEnrollment, 0, len(res))
	for _, e := range res {
		out = append(out, toPracticeEnrollmentResp(e))
	}
	c.JSON(http.StatusOK, out)
}

func toPracticeResp(p m.Practice) resp.Practice {
	return resp.Practice{ID: p.ID, Title: p.Title, Description: p.Description, DepartmentID: p.DepartmentID, ProgramID: p.ProgramID, StartDate: p.StartDate, EndDate: p.EndDate, SupervisorID: p.SupervisorID}
}

func toPracticeEnrollmentResp(e m.PracticeEnrollment) resp.PracticeEnrollment {
	return resp.PracticeEnrollment{ID: e.ID, PracticeID: e.PracticeID, StudentID: e.StudentID, Place: e.Place, Status: e.Status}
}
