package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	reqdto "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/service"
)

// AdminController manages administrator endpoints.
type AdminController struct {
	admins *service.AdminService
}

// NewAdminController creates the controller.
func NewAdminController(admins *service.AdminService) *AdminController {
	return &AdminController{admins: admins}
}

// RegisterRoutes registers admin endpoints.
func (c *AdminController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/deans", c.createDean)
	rg.GET("/deans", c.listDeans)
}

// createDean godoc
// @Summary Create dean staff
// @Security BearerAuth
// @Tags Admin
// @Accept json
// @Produce json
// @Param payload body request.CreateDeanRequest true "Dean payload"
// @Success 201 {object} response.UserProfile
// @Failure 400 {object} gin.H
// @Router /admin/deans [post]
func (c *AdminController) createDean(ctx *gin.Context) {
	var payload reqdto.CreateDeanRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.admins.CreateDean(ctx.Request.Context(), payload)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

// listDeans godoc
// @Summary List dean staff
// @Security BearerAuth
// @Tags Admin
// @Produce json
// @Success 200 {array} response.UserProfile
// @Router /admin/deans [get]
func (c *AdminController) listDeans(ctx *gin.Context) {
	resp, err := c.admins.ListDeans(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
