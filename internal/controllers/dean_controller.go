package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	reqdto "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/service"
)

// DeanController exposes dean office endpoints.
type DeanController struct {
	deans *service.DeanService
}

// NewDeanController builds controller.
func NewDeanController(deans *service.DeanService) *DeanController {
	return &DeanController{deans: deans}
}

// RegisterRoutes wires dean routes.
func (c *DeanController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/groups", c.createGroup)
	rg.GET("/groups", c.listGroups)
	rg.POST("/subjects", c.createSubject)
	rg.GET("/subjects", c.listSubjects)
	rg.POST("/teachers", c.createTeacher)
	rg.GET("/teachers", c.listTeachers)
	rg.POST("/students", c.createStudent)
	rg.GET("/students", c.listStudents)
	rg.POST("/subjects/:subjectId/assign", c.assignTeacher)
	rg.POST("/subjects/:subjectId/groups", c.attachGroup)
	rg.POST("/groups/:groupId/students", c.assignStudentToGroup)
	rg.POST("/sessions", c.scheduleSession)
	rg.GET("/groups/ranking", c.groupRanking)
}

func (c *DeanController) createGroup(ctx *gin.Context) {
	var payload reqdto.CreateGroupRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.deans.CreateGroup(ctx.Request.Context(), payload)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

func (c *DeanController) listGroups(ctx *gin.Context) {
	resp, err := c.deans.ListGroups(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *DeanController) createSubject(ctx *gin.Context) {
	var payload reqdto.CreateSubjectRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.deans.CreateSubject(ctx.Request.Context(), payload)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

func (c *DeanController) listSubjects(ctx *gin.Context) {
	resp, err := c.deans.ListSubjects(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *DeanController) createTeacher(ctx *gin.Context) {
	var payload reqdto.CreateTeacherRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.deans.CreateTeacher(ctx.Request.Context(), payload)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

func (c *DeanController) listTeachers(ctx *gin.Context) {
	resp, err := c.deans.ListTeachers(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *DeanController) createStudent(ctx *gin.Context) {
	var payload reqdto.CreateStudentRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.deans.CreateStudent(ctx.Request.Context(), payload)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

func (c *DeanController) listStudents(ctx *gin.Context) {
	resp, err := c.deans.ListStudents(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *DeanController) assignTeacher(ctx *gin.Context) {
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid subject id"})
		return
	}
	var payload reqdto.AssignTeacherRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.deans.AssignTeacher(ctx.Request.Context(), subjectID, payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *DeanController) attachGroup(ctx *gin.Context) {
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid subject id"})
		return
	}
	var payload reqdto.AttachGroupRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.deans.AttachGroup(ctx.Request.Context(), subjectID, payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *DeanController) assignStudentToGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var payload reqdto.AssignStudentToGroupRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.deans.AssignStudentToGroup(ctx.Request.Context(), groupID, payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *DeanController) scheduleSession(ctx *gin.Context) {
	var payload reqdto.ScheduleSessionRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.deans.ScheduleSession(ctx.Request.Context(), payload)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

func (c *DeanController) groupRanking(ctx *gin.Context) {
	resp, err := c.deans.GroupRanking(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
