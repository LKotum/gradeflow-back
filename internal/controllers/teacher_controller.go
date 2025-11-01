package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	reqdto "gradeflow/internal/domain/dto/request"
	response "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/middleware"
	"gradeflow/internal/service"
	"gradeflow/pkg/httpx"
)

// TeacherController exposes teacher operations.
type TeacherController struct {
	teachers *service.TeacherService
}

// NewTeacherController builds controller.
func NewTeacherController(teachers *service.TeacherService) *TeacherController {
	return &TeacherController{teachers: teachers}
}

var (
	_ response.TeacherDashboardResponse
	_ []response.ScheduleEntry
)

// RegisterRoutes registers teacher endpoints.
// @Summary Teacher operations
// @Tags Teacher
// @Security BearerAuth
// @BasePath /teacher
func (c *TeacherController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/dashboard", c.dashboard)
	rg.GET("/subjects/:subjectId/groups/:groupId/grades", c.gradeTable)
	rg.POST("/grades", c.upsertGrade)
	rg.PATCH("/grades/:gradeId", c.updateGrade)
	rg.DELETE("/grades/:gradeId", c.deleteGrade)
	rg.GET("/subjects/:subjectId/groups/:groupId/students/:studentId/averages", c.averages)
	rg.GET("/schedule", c.schedule)
}

func (c *TeacherController) teacherID(ctx *gin.Context) (uuid.UUID, bool) {
	val, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		httpx.WriteError(ctx, http.StatusUnauthorized, "missing_teacher", "teacher context missing", nil)
		return uuid.UUID{}, false
	}
	idStr, ok := val.(string)
	if !ok {
		httpx.WriteError(ctx, http.StatusUnauthorized, "invalid_context", "invalid teacher context", nil)
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(ctx, http.StatusUnauthorized, "invalid_context", "invalid teacher identifier", nil)
		return uuid.UUID{}, false
	}
	return id, true
}

// dashboard godoc
// @Summary      Кабинет преподавателя
// @Security     BearerAuth
// @Tags         Teacher
// @Produce      json
// @Success      200 {object} response.TeacherDashboardResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /teacher/dashboard [get]
func (c *TeacherController) dashboard(ctx *gin.Context) {
	teacherID, ok := c.teacherID(ctx)
	if !ok {
		return
	}
	resp, err := c.teachers.Dashboard(ctx.Request.Context(), teacherID)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "teacher_dashboard_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// gradeTable godoc
// @Summary      Таблица оценок по предмету и группе
// @Security     BearerAuth
// @Tags         Teacher
// @Produce      json
// @Param        subjectId path string true "ID предмета"
// @Param        groupId   path string true "ID группы"
// @Param        limit     query int    false "Лимит студентов"
// @Param        offset    query int    false "Смещение"
// @Param        search    query string false "Поиск по ФИО или ИНС"
// @Success      200 {object} response.GradeTableResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /teacher/subjects/{subjectId}/groups/{groupId}/grades [get]
func (c *TeacherController) gradeTable(ctx *gin.Context) {
	teacherID, ok := c.teacherID(ctx)
	if !ok {
		return
	}
	var query reqdto.PaginationQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_subject", "invalid subject id", nil)
		return
	}
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_group", "invalid group id", nil)
		return
	}
	resp, err := c.teachers.GradeTable(ctx.Request.Context(), teacherID, subjectID, groupID)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "grade_table_failed", err.Error(), nil)
		return
	}
	filteredStudents := resp.Students
	if query.Search != nil && strings.TrimSpace(*query.Search) != "" {
		needle := strings.ToLower(strings.TrimSpace(*query.Search))
		tmp := make([]response.UserProfile, 0, len(filteredStudents))
		for _, student := range filteredStudents {
			fullName := strings.ToLower(strings.Join([]string{student.LastName, student.FirstName, deref(student.MiddleName)}, " "))
			if strings.Contains(fullName, needle) ||
				(student.INS != nil && strings.Contains(strings.ToLower(*student.INS), needle)) {
				tmp = append(tmp, student)
			}
		}
		filteredStudents = tmp
	}
	total := len(filteredStudents)
	query.Normalize(100)
	if total == 0 {
		resp.Students = []response.UserProfile{}
		resp.Grades = []response.GradeDetail{}
		resp.Meta = &response.PageMeta{Limit: query.Limit, Offset: 0, Total: total}
		httpx.WriteData(ctx, http.StatusOK, resp)
		return
	}
	if query.Offset > total {
		query.Offset = total
	}
	end := query.Offset + query.Limit
	if end > total {
		end = total
	}
	pagedStudents := filteredStudents[query.Offset:end]
	allowed := make(map[string]struct{}, len(pagedStudents))
	for _, student := range pagedStudents {
		allowed[student.ID] = struct{}{}
	}
	filteredGrades := make([]response.GradeDetail, 0, len(resp.Grades))
	for _, grade := range resp.Grades {
		if _, ok := allowed[grade.Student.ID]; ok {
			filteredGrades = append(filteredGrades, grade)
		}
	}
	resp.Students = pagedStudents
	resp.Grades = filteredGrades
	resp.Meta = &response.PageMeta{Limit: query.Limit, Offset: query.Offset, Total: total}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// upsertGrade godoc
// @Summary      Выставить или обновить оценку
// @Security     BearerAuth
// @Tags         Teacher
// @Accept       json
// @Produce      json
// @Param        payload body request.GradeUpsertRequest true "Данные оценки"
// @Success      201 {object} response.GradeDetail
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /teacher/grades [post]
func (c *TeacherController) upsertGrade(ctx *gin.Context) {
	teacherID, ok := c.teacherID(ctx)
	if !ok {
		return
	}
	var payload reqdto.GradeUpsertRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.teachers.UpsertGrade(ctx.Request.Context(), teacherID, payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "grade_upsert_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusCreated, resp)
}

// updateGrade godoc
// @Summary      Изменить оценку
// @Security     BearerAuth
// @Tags         Teacher
// @Accept       json
// @Produce      json
// @Param        gradeId path string true "ID оценки"
// @Param        payload body request.GradeUpdateRequest true "Новая оценка"
// @Success      200 {object} response.GradeDetail
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /teacher/grades/{gradeId} [patch]
func (c *TeacherController) updateGrade(ctx *gin.Context) {
	teacherID, ok := c.teacherID(ctx)
	if !ok {
		return
	}
	gradeID, err := uuid.Parse(ctx.Param("gradeId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_grade", "invalid grade id", nil)
		return
	}
	var payload reqdto.GradeUpdateRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.teachers.UpdateGrade(ctx.Request.Context(), teacherID, gradeID, payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "grade_update_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// deleteGrade godoc
// @Summary      Удалить оценку
// @Security     BearerAuth
// @Tags         Teacher
// @Param        gradeId path string true "ID оценки"
// @Success      204
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /teacher/grades/{gradeId} [delete]
func (c *TeacherController) deleteGrade(ctx *gin.Context) {
	teacherID, ok := c.teacherID(ctx)
	if !ok {
		return
	}
	gradeID, err := uuid.Parse(ctx.Param("gradeId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_grade", "invalid grade id", nil)
		return
	}
	if err := c.teachers.DeleteGrade(ctx.Request.Context(), teacherID, gradeID); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "grade_delete_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// averages godoc
// @Summary      Средние показатели студента
// @Security     BearerAuth
// @Tags         Teacher
// @Produce      json
// @Param        subjectId path string true "ID предмета"
// @Param        groupId   path string true "ID группы"
// @Param        studentId path string true "ID студента"
// @Success      200 {object} response.AverageMetricResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /teacher/subjects/{subjectId}/groups/{groupId}/students/{studentId}/averages [get]
func (c *TeacherController) averages(ctx *gin.Context) {
	if _, ok := c.teacherID(ctx); !ok {
		return
	}
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_subject", "invalid subject id", nil)
		return
	}
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
	resp, err := c.teachers.SubjectAverages(ctx.Request.Context(), groupID, subjectID, studentID)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "subject_average_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// schedule godoc
// @Summary      Расписание преподавателя
// @Security     BearerAuth
// @Tags         Teacher
// @Produce      json
// @Param        subjectId query string false "ID предмета"
// @Param        groupId   query string false "ID группы"
// @Param        from      query string false "Дата с" format(date)
// @Param        to        query string false "Дата по" format(date)
// @Success      200 {array} response.ScheduleEntry
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /teacher/schedule [get]
func (c *TeacherController) schedule(ctx *gin.Context) {
	teacherID, ok := c.teacherID(ctx)
	if !ok {
		return
	}
	var query reqdto.ScheduleQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	entries, err := c.teachers.Schedule(ctx.Request.Context(), teacherID, query)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "schedule_fetch_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, entries)
}

func deref(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}
