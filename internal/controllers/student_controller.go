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

// StudentController exposes student endpoints.
type StudentController struct {
	students *service.StudentService
}

var (
	_ response.StudentDashboardResponse
	_ response.StudentSubjectGrade
	_ response.AverageMetricResponse
	_ []response.ScheduleEntry
)

// NewStudentController creates controller.
func NewStudentController(students *service.StudentService) *StudentController {
	return &StudentController{students: students}
}

// RegisterRoutes wires student endpoints.
func (c *StudentController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/dashboard", c.dashboard)
	rg.GET("/subjects", c.subjects)
	rg.GET("/subjects/:subjectId/averages", c.subjectAverage)
	rg.GET("/schedule", c.schedule)
}

// dashboard godoc
// @Summary      Личный кабинет студента
// @Security     BearerAuth
// @Tags         Student
// @Produce      json
// @Success      200 {object} response.StudentDashboardResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /student/dashboard [get]
func (c *StudentController) studentID(ctx *gin.Context) (uuid.UUID, bool) {
	val, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		httpx.WriteError(ctx, http.StatusUnauthorized, "missing_student", "student context missing", nil)
		return uuid.UUID{}, false
	}
	idStr, ok := val.(string)
	if !ok {
		httpx.WriteError(ctx, http.StatusUnauthorized, "invalid_context", "invalid student context", nil)
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(ctx, http.StatusUnauthorized, "invalid_context", "invalid student identifier", nil)
		return uuid.UUID{}, false
	}
	return id, true
}

// dashboard godoc
// @Summary      Личный кабинет студента
// @Security     BearerAuth
// @Tags         Student
// @Produce      json
// @Success      200 {object} response.StudentDashboardResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /student/dashboard [get]
func (c *StudentController) dashboard(ctx *gin.Context) {
	studentID, ok := c.studentID(ctx)
	if !ok {
		return
	}
	resp, err := c.students.Dashboard(ctx.Request.Context(), studentID)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "student_dashboard_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// subjects godoc
// @Summary      Список предметов студента
// @Security     BearerAuth
// @Tags         Student
// @Produce      json
// @Param        limit  query int    false "Лимит"
// @Param        offset query int    false "Смещение"
// @Param        search query string false "Поиск по названию или коду"
// @Success      200 {object} response.PaginatedStudentSubjects
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /student/subjects [get]
func (c *StudentController) subjects(ctx *gin.Context) {
	var query reqdto.PaginationQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	studentID, ok := c.studentID(ctx)
	if !ok {
		return
	}
	resp, err := c.students.Subjects(ctx.Request.Context(), studentID)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "student_subjects_failed", err.Error(), nil)
		return
	}
	filtered := resp
	if query.Search != nil && strings.TrimSpace(*query.Search) != "" {
		needle := strings.ToLower(strings.TrimSpace(*query.Search))
		tmp := make([]response.StudentSubjectGrade, 0, len(filtered))
		for _, subj := range filtered {
			if strings.Contains(strings.ToLower(subj.Subject.Name), needle) ||
				strings.Contains(strings.ToLower(subj.Subject.Code), needle) {
				tmp = append(tmp, subj)
			}
		}
		filtered = tmp
	}
	total := len(filtered)
	query.Normalize(100)
	if total == 0 {
		result := response.Paginated[response.StudentSubjectGrade]{
			Data: []response.StudentSubjectGrade{},
			Meta: response.PageMeta{Limit: query.Limit, Offset: 0, Total: 0},
		}
		httpx.WriteData(ctx, http.StatusOK, result)
		return
	}
	if query.Offset > total {
		query.Offset = total
	}
	end := query.Offset + query.Limit
	if end > total {
		end = total
	}
	paged := filtered[query.Offset:end]
	result := response.Paginated[response.StudentSubjectGrade]{
		Data: paged,
		Meta: response.PageMeta{Limit: query.Limit, Offset: query.Offset, Total: total},
	}
	httpx.WriteData(ctx, http.StatusOK, result)
}

// subjectAverage godoc
// @Summary      Средний балл по предмету
// @Security     BearerAuth
// @Tags         Student
// @Produce      json
// @Param        subjectId path string true "ID предмета"
// @Success      200 {object} response.AverageMetricResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /student/subjects/{subjectId}/averages [get]
func (c *StudentController) subjectAverage(ctx *gin.Context) {
	studentID, ok := c.studentID(ctx)
	if !ok {
		return
	}
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_subject", "invalid subject id", nil)
		return
	}
	resp, err := c.students.SubjectAverage(ctx.Request.Context(), studentID, subjectID)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "student_average_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// schedule godoc
// @Summary      Расписание студента
// @Security     BearerAuth
// @Tags         Student
// @Produce      json
// @Param        subjectId query string false "ID предмета"
// @Param        from query string false "Дата с" format(date)
// @Param        to   query string false "Дата по" format(date)
// @Success      200 {array} response.ScheduleEntry
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /student/schedule [get]
func (c *StudentController) schedule(ctx *gin.Context) {
	studentID, ok := c.studentID(ctx)
	if !ok {
		return
	}
	var query reqdto.ScheduleQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	entries, err := c.students.Schedule(ctx.Request.Context(), studentID, query)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "schedule_fetch_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, entries)
}
