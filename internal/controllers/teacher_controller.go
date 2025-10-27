package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	reqdto "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/middleware"
	"gradeflow/internal/service"
)

// TeacherController exposes teacher operations.
type TeacherController struct {
	teachers *service.TeacherService
}

// NewTeacherController builds controller.
func NewTeacherController(teachers *service.TeacherService) *TeacherController {
	return &TeacherController{teachers: teachers}
}

// RegisterRoutes registers teacher endpoints.
func (c *TeacherController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/dashboard", c.dashboard)
	rg.GET("/subjects/:subjectId/groups/:groupId/grades", c.gradeTable)
	rg.POST("/grades", c.upsertGrade)
	rg.PATCH("/grades/:gradeId", c.updateGrade)
	rg.DELETE("/grades/:gradeId", c.deleteGrade)
	rg.GET("/subjects/:subjectId/groups/:groupId/students/:studentId/averages", c.averages)
}

func (c *TeacherController) teacherID(ctx *gin.Context) (uuid.UUID, bool) {
	val, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing teacher"})
		return uuid.UUID{}, false
	}
	idStr, ok := val.(string)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid teacher"})
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid teacher"})
		return uuid.UUID{}, false
	}
	return id, true
}

func (c *TeacherController) dashboard(ctx *gin.Context) {
	teacherID, ok := c.teacherID(ctx)
	if !ok {
		return
	}
	resp, err := c.teachers.Dashboard(ctx.Request.Context(), teacherID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *TeacherController) gradeTable(ctx *gin.Context) {
	teacherID, ok := c.teacherID(ctx)
	if !ok {
		return
	}
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid subject id"})
		return
	}
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	resp, err := c.teachers.GradeTable(ctx.Request.Context(), teacherID, subjectID, groupID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *TeacherController) upsertGrade(ctx *gin.Context) {
	teacherID, ok := c.teacherID(ctx)
	if !ok {
		return
	}
	var payload reqdto.GradeUpsertRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.teachers.UpsertGrade(ctx.Request.Context(), teacherID, payload)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

func (c *TeacherController) updateGrade(ctx *gin.Context) {
	teacherID, ok := c.teacherID(ctx)
	if !ok {
		return
	}
	gradeID, err := uuid.Parse(ctx.Param("gradeId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid grade id"})
		return
	}
	var payload reqdto.GradeUpdateRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.teachers.UpdateGrade(ctx.Request.Context(), teacherID, gradeID, payload)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *TeacherController) deleteGrade(ctx *gin.Context) {
	teacherID, ok := c.teacherID(ctx)
	if !ok {
		return
	}
	gradeID, err := uuid.Parse(ctx.Param("gradeId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid grade id"})
		return
	}
	if err := c.teachers.DeleteGrade(ctx.Request.Context(), teacherID, gradeID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *TeacherController) averages(ctx *gin.Context) {
	if _, ok := c.teacherID(ctx); !ok {
		return
	}
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid subject id"})
		return
	}
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	studentID, err := uuid.Parse(ctx.Param("studentId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}
	resp, err := c.teachers.SubjectAverages(ctx.Request.Context(), groupID, subjectID, studentID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
