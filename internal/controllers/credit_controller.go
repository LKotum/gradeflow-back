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

type CreditController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewCreditController(db *gorm.DB, cfg config.Config) *CreditController {
	return &CreditController{DB: db, Cfg: cfg}
}

func (h *CreditController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/credits")
	g.POST("", middleware.AdminOrDean(), h.create)
	g.GET("", h.list)
	g.GET(":id", h.get)
	g.PUT(":id", middleware.AdminOrDean(), h.update)
	g.DELETE(":id", middleware.AdminOrDean(), h.delete)
}

// @Summary Create credit/offset
// @Tags credits
// @Accept json
// @Produce json
// @Param input body request.CreateCredit true "credit"
// @Success 201 {object} response.Credit
// @Failure 400 {object} response.Error
// @Router /credits [post]
func (h *CreditController) create(c *gin.Context) {
	var in req.CreateCredit
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	if in.Type != "credit" && in.Type != "offset" && in.Type != "transfer" {
		c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid type"})
		return
	}
	cr := m.Credit{CourseID: in.CourseID, StudentID: in.StudentID, Type: in.Type, Reason: in.Reason, ApprovedByID: in.ApprovedByID}
	if err := h.DB.Create(&cr).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toCreditResp(cr))
}

// @Summary List credits
// @Tags credits
// @Produce json
// @Param courseId query string false "filter by course"
// @Param studentId query string false "filter by student"
// @Param type query string false "filter by type"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} response.CreditList
// @Router /credits [get]
func (h *CreditController) list(c *gin.Context) {
	var qin req.ListCreditQuery
	_ = c.ShouldBindQuery(&qin)
	base := h.DB.Model(&m.Credit{})
	if v := qin.CourseID; v != "" {
		base = base.Where("course_id = ?", v)
	}
	if v := qin.StudentID; v != "" {
		base = base.Where("student_id = ?", v)
	}
	if v := qin.Type; v != "" {
		base = base.Where("type = ?", v)
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
	var ll []m.Credit
	if err := base.Order("created_at desc").Limit(qin.Limit).Offset(qin.Offset).Find(&ll).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	items := make([]resp.Credit, 0, len(ll))
	for _, cr := range ll {
		items = append(items, toCreditResp(cr))
	}
	c.JSON(http.StatusOK, resp.List[resp.Credit]{Items: items, Page: resp.Page{Limit: qin.Limit, Offset: qin.Offset, Total: total}})
}

// @Summary Get credit
// @Tags credits
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} response.Credit
// @Failure 404 {object} response.Error
// @Router /credits/{id} [get]
func (h *CreditController) get(c *gin.Context) {
	id := c.Param("id")
	var cr m.Credit
	if err := h.DB.First(&cr, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, toCreditResp(cr))
}

// @Summary Update credit
// @Tags credits
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param input body request.UpdateCredit true "credit"
// @Success 200 {object} response.Credit
// @Failure 400 {object} response.Error
// @Router /credits/{id} [put]
func (h *CreditController) update(c *gin.Context) {
	id := c.Param("id")
	var in req.UpdateCredit
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var cr m.Credit
	if err := h.DB.First(&cr, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, resp.Error{Error: "not found"})
		return
	}
	if in.Type != nil {
		if *in.Type != "credit" && *in.Type != "offset" && *in.Type != "transfer" {
			c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid type"})
			return
		}
		cr.Type = *in.Type
	}
	if in.Reason != nil {
		cr.Reason = *in.Reason
	}
	if in.ApprovedByID != nil {
		cr.ApprovedByID = in.ApprovedByID
	}
	if err := h.DB.Save(&cr).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toCreditResp(cr))
}

// @Summary Delete credit
// @Tags credits
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} map[string]bool
// @Router /credits/{id} [delete]
func (h *CreditController) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&m.Credit{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func toCreditResp(cr m.Credit) resp.Credit {
	return resp.Credit{ID: cr.ID, CourseID: cr.CourseID, StudentID: cr.StudentID, Type: cr.Type, Reason: cr.Reason, ApprovedByID: cr.ApprovedByID}
}
