package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	reqdto "gradeflow/internal/domain/dto/request"
	response "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/service"
	"gradeflow/pkg/httpx"
)

var (
	_ response.GroupSummary
	_ response.SubjectSummary
	_ response.UserProfile
	_ response.SessionSummary
	_ response.GroupRankingResponse
	_ response.PaginatedGroupSummaries
	_ response.PaginatedSubjectSummaries
	_ response.PaginatedUserProfiles
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
	rg.PATCH("/groups/:groupId", c.updateGroup)
	rg.DELETE("/groups/:groupId", c.deleteGroup)
	rg.POST("/subjects", c.createSubject)
	rg.GET("/subjects", c.listSubjects)
	rg.PATCH("/subjects/:subjectId", c.updateSubject)
	rg.DELETE("/subjects/:subjectId", c.deleteSubject)
	rg.POST("/teachers", c.createTeacher)
	rg.GET("/teachers", c.listTeachers)
	rg.PATCH("/teachers/:teacherId", c.updateTeacher)
	rg.POST("/students", c.createStudent)
	rg.GET("/students", c.listStudents)
	rg.PATCH("/students/:studentId", c.updateStudent)
	rg.POST("/subjects/:subjectId/assign", c.assignTeacher)
	rg.POST("/subjects/:subjectId/groups", c.attachGroup)
	rg.POST("/groups/:groupId/students", c.assignStudentToGroup)
	rg.DELETE("/groups/:groupId/students/:studentId", c.detachStudentFromGroup)
	rg.DELETE("/subjects/:subjectId/teachers/:teacherId", c.detachTeacherFromSubject)
	rg.GET("/schedule", c.schedule)
	rg.POST("/sessions", c.scheduleSession)
	rg.GET("/groups/ranking", c.groupRanking)
}

// createGroup godoc
// @Summary Create group
// @Security BearerAuth
// @Tags Dean
// @Accept json
// @Produce json
// @Param payload body request.CreateGroupRequest true "Group payload"
// @Success 201 {object} response.GroupSummary
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/groups [post]
func (c *DeanController) createGroup(ctx *gin.Context) {
	var payload reqdto.CreateGroupRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.deans.CreateGroup(ctx.Request.Context(), payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "group_create_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusCreated, resp)
}

// listGroups godoc
// @Summary List groups
// @Security BearerAuth
// @Tags Dean
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search phrase"
// @Success 200 {object} response.PaginatedGroupSummaries
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /dean/groups [get]
func (c *DeanController) listGroups(ctx *gin.Context) {
	var query reqdto.PaginationQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	resp, err := c.deans.ListGroups(ctx.Request.Context(), query)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "group_list_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// updateGroup godoc
// @Summary Update group
// @Security BearerAuth
// @Tags Dean
// @Accept json
// @Produce json
// @Param groupId path string true "Group ID"
// @Param payload body request.UpdateGroupRequest true "Update payload"
// @Success 200 {object} response.GroupSummary
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/groups/{groupId} [patch]
func (c *DeanController) updateGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_group", "invalid group id", nil)
		return
	}
	var payload reqdto.UpdateGroupRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.deans.UpdateGroup(ctx.Request.Context(), groupID, payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "group_update_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// deleteGroup godoc
// @Summary Soft delete group
// @Security BearerAuth
// @Tags Dean
// @Param groupId path string true "Group ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/groups/{groupId} [delete]
func (c *DeanController) deleteGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_group", "invalid group id", nil)
		return
	}
	if err := c.deans.DeleteGroup(ctx.Request.Context(), groupID); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "group_delete_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// createSubject godoc
// @Summary Create subject
// @Security BearerAuth
// @Tags Dean
// @Accept json
// @Produce json
// @Param payload body request.CreateSubjectRequest true "Subject payload"
// @Success 201 {object} response.SubjectSummary
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/subjects [post]
func (c *DeanController) createSubject(ctx *gin.Context) {
	var payload reqdto.CreateSubjectRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.deans.CreateSubject(ctx.Request.Context(), payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "subject_create_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusCreated, resp)
}

// listSubjects godoc
// @Summary List subjects
// @Security BearerAuth
// @Tags Dean
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search phrase"
// @Success 200 {object} response.PaginatedSubjectSummaries
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /dean/subjects [get]
func (c *DeanController) listSubjects(ctx *gin.Context) {
	var query reqdto.PaginationQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	resp, err := c.deans.ListSubjects(ctx.Request.Context(), query)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "subject_list_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// updateSubject godoc
// @Summary Update subject
// @Security BearerAuth
// @Tags Dean
// @Accept json
// @Produce json
// @Param subjectId path string true "Subject ID"
// @Param payload body request.UpdateSubjectRequest true "Update payload"
// @Success 200 {object} response.SubjectSummary
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/subjects/{subjectId} [patch]
func (c *DeanController) updateSubject(ctx *gin.Context) {
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_subject", "invalid subject id", nil)
		return
	}
	var payload reqdto.UpdateSubjectRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.deans.UpdateSubject(ctx.Request.Context(), subjectID, payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "subject_update_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// deleteSubject godoc
// @Summary Soft delete subject
// @Security BearerAuth
// @Tags Dean
// @Param subjectId path string true "Subject ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/subjects/{subjectId} [delete]
func (c *DeanController) deleteSubject(ctx *gin.Context) {
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_subject", "invalid subject id", nil)
		return
	}
	if err := c.deans.DeleteSubject(ctx.Request.Context(), subjectID); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "subject_delete_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// createTeacher godoc
// @Summary Create teacher
// @Security BearerAuth
// @Tags Dean
// @Accept json
// @Produce json
// @Param payload body request.CreateTeacherRequest true "Teacher payload"
// @Success 201 {object} response.UserProfile
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/teachers [post]
func (c *DeanController) createTeacher(ctx *gin.Context) {
	var payload reqdto.CreateTeacherRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.deans.CreateTeacher(ctx.Request.Context(), payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "teacher_create_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusCreated, resp)
}

// listTeachers godoc
// @Summary List teachers
// @Security BearerAuth
// @Tags Dean
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search phrase"
// @Success 200 {object} response.PaginatedUserProfiles
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /dean/teachers [get]
func (c *DeanController) listTeachers(ctx *gin.Context) {
	var query reqdto.PaginationQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	resp, err := c.deans.ListTeachers(ctx.Request.Context(), query)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "teacher_list_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// updateTeacher godoc
// @Summary Update teacher
// @Security BearerAuth
// @Tags Dean
// @Accept json
// @Produce json
// @Param teacherId path string true "Teacher ID"
// @Param payload body request.UpdateTeacherRequest true "Update payload"
// @Success 200 {object} response.UserProfile
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/teachers/{teacherId} [patch]
func (c *DeanController) updateTeacher(ctx *gin.Context) {
	teacherID, err := uuid.Parse(ctx.Param("teacherId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_teacher", "invalid teacher id", nil)
		return
	}
	var payload reqdto.UpdateTeacherRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.deans.UpdateTeacher(ctx.Request.Context(), teacherID, payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "teacher_update_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// createStudent godoc
// @Summary Create student
// @Security BearerAuth
// @Tags Dean
// @Accept json
// @Produce json
// @Param payload body request.CreateStudentRequest true "Student payload"
// @Success 201 {object} response.UserProfile
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/students [post]
func (c *DeanController) createStudent(ctx *gin.Context) {
	var payload reqdto.CreateStudentRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.deans.CreateStudent(ctx.Request.Context(), payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "student_create_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusCreated, resp)
}

// listStudents godoc
// @Summary List students
// @Security BearerAuth
// @Tags Dean
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search phrase"
// @Success 200 {object} response.PaginatedUserProfiles
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /dean/students [get]
func (c *DeanController) listStudents(ctx *gin.Context) {
	var query reqdto.PaginationQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	resp, err := c.deans.ListStudents(ctx.Request.Context(), query)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "student_list_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// updateStudent godoc
// @Summary Update student
// @Security BearerAuth
// @Tags Dean
// @Accept json
// @Produce json
// @Param studentId path string true "Student ID"
// @Param payload body request.UpdateStudentRequest true "Update payload"
// @Success 200 {object} response.UserProfile
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/students/{studentId} [patch]
func (c *DeanController) updateStudent(ctx *gin.Context) {
	studentID, err := uuid.Parse(ctx.Param("studentId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_student", "invalid student id", nil)
		return
	}
	var payload reqdto.UpdateStudentRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.deans.UpdateStudent(ctx.Request.Context(), studentID, payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "student_update_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// assignTeacher godoc
// @Summary Assign teacher to subject
// @Security BearerAuth
// @Tags Dean
// @Accept json
// @Param subjectId path string true "Subject ID"
// @Param payload body request.AssignTeacherRequest true "Assignment payload"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/subjects/{subjectId}/assign [post]
func (c *DeanController) assignTeacher(ctx *gin.Context) {
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_subject", "invalid subject id", nil)
		return
	}
	var payload reqdto.AssignTeacherRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	if err := c.deans.AssignTeacher(ctx.Request.Context(), subjectID, payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "assign_teacher_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// attachGroup godoc
// @Summary Attach group to subject
// @Security BearerAuth
// @Tags Dean
// @Accept json
// @Param subjectId path string true "Subject ID"
// @Param payload body request.AttachGroupRequest true "Group link payload"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/subjects/{subjectId}/groups [post]
func (c *DeanController) attachGroup(ctx *gin.Context) {
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_subject", "invalid subject id", nil)
		return
	}
	var payload reqdto.AttachGroupRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	if err := c.deans.AttachGroup(ctx.Request.Context(), subjectID, payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "attach_group_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// assignStudentToGroup godoc
// @Summary Назначить студентов в группу
// @Security BearerAuth
// @Tags Dean
// @Accept json
// @Param groupId path string true "ID группы"
// @Param payload body request.AssignStudentToGroupRequest true "Список студентов"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/groups/{groupId}/students [post]
func (c *DeanController) assignStudentToGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_group", "invalid group id", nil)
		return
	}
	var payload reqdto.AssignStudentToGroupRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	if err := c.deans.AssignStudentToGroup(ctx.Request.Context(), groupID, payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "assign_student_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// detachStudentFromGroup godoc
// @Summary Открепить студента от группы
// @Security BearerAuth
// @Tags Dean
// @Param groupId path string true "ID группы"
// @Param studentId path string true "ID студента"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/groups/{groupId}/students/{studentId} [delete]
func (c *DeanController) detachStudentFromGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_group", "invalid group id", nil)
		return
	}
	studentID, err := uuid.Parse(ctx.Param("studentId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_student", "invalid student id", nil)
		return
	}
	if err := c.deans.DetachStudentFromGroup(ctx.Request.Context(), groupID, studentID); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "detach_student_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// detachTeacherFromSubject godoc
// @Summary Открепить преподавателя от предмета
// @Security BearerAuth
// @Tags Dean
// @Param subjectId path string true "ID предмета"
// @Param teacherId path string true "ID преподавателя"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/subjects/{subjectId}/teachers/{teacherId} [delete]
func (c *DeanController) detachTeacherFromSubject(ctx *gin.Context) {
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_subject", "invalid subject id", nil)
		return
	}
	teacherID, err := uuid.Parse(ctx.Param("teacherId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_teacher", "invalid teacher id", nil)
		return
	}
	if err := c.deans.DetachTeacherFromSubject(ctx.Request.Context(), subjectID, teacherID); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "detach_teacher_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// scheduleSession godoc
// @Summary Создать занятие
// @Security BearerAuth
// @Tags Dean
// @Accept json
// @Produce json
// @Param payload body request.ScheduleSessionRequest true "Пара для предмета"
// @Success 201 {array} response.SessionSummary
// @Failure 400 {object} response.ErrorResponse
// @Router /dean/sessions [post]
func (c *DeanController) scheduleSession(ctx *gin.Context) {
	var payload reqdto.ScheduleSessionRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.deans.ScheduleSession(ctx.Request.Context(), payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "session_create_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusCreated, resp)
}

// schedule godoc
// @Summary Расписание занятий
// @Security BearerAuth
// @Tags Dean
// @Produce json
// @Param subjectId query string false "ID предмета"
// @Param groupId query string false "ID группы"
// @Param teacherId query string false "ID преподавателя"
// @Param from query string false "Дата с" format(date)
// @Param to query string false "Дата по" format(date)
// @Success 200 {array} response.ScheduleEntry
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /dean/schedule [get]
func (c *DeanController) schedule(ctx *gin.Context) {
	var query reqdto.ScheduleQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	entries, err := c.deans.Schedule(ctx.Request.Context(), query)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "schedule_fetch_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, entries)
}

// groupRanking godoc
// @Summary Group ranking by average grade
// @Security BearerAuth
// @Tags Dean
// @Produce json
// @Success 200 {object} response.GroupRankingResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /dean/groups/ranking [get]
func (c *DeanController) groupRanking(ctx *gin.Context) {
	resp, err := c.deans.GroupRanking(ctx.Request.Context())
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "group_ranking_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}
