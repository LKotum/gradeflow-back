package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gradeflow/internal/middleware"
	"gradeflow/internal/service"
)

// StudentController exposes student endpoints.
type StudentController struct {
	students *service.StudentService
}

// NewStudentController creates controller.
func NewStudentController(students *service.StudentService) *StudentController {
	return &StudentController{students: students}
}

// RegisterRoutes wires student endpoints.
func (c *StudentController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/dashboard", c.dashboard)
	rg.GET("/subjects", c.subjects)
	rg.GET("/subjects/:subjectId/averages", c.subjectAverage)
}

func (c *StudentController) studentID(ctx *gin.Context) (uuid.UUID, bool) {
	val, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing student"})
		return uuid.UUID{}, false
	}
	idStr, ok := val.(string)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid student"})
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid student id"})
		return uuid.UUID{}, false
	}
	return id, true
}

func (c *StudentController) dashboard(ctx *gin.Context) {
	studentID, ok := c.studentID(ctx)
	if !ok {
		return
	}
	resp, err := c.students.Dashboard(ctx.Request.Context(), studentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *StudentController) subjects(ctx *gin.Context) {
	studentID, ok := c.studentID(ctx)
	if !ok {
		return
	}
	resp, err := c.students.Subjects(ctx.Request.Context(), studentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *StudentController) subjectAverage(ctx *gin.Context) {
	studentID, ok := c.studentID(ctx)
	if !ok {
		return
	}
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid subject id"})
		return
	}
	resp, err := c.students.SubjectAverage(ctx.Request.Context(), studentID, subjectID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
