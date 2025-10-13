package controllers

import (
    "net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gradeflow/internal/config"
	resp "gradeflow/internal/domain/dto/response"
	m "gradeflow/internal/domain/models"
    "gradeflow/pkg/middleware"
)

type ReportController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewReportController(db *gorm.DB, cfg config.Config) *ReportController {
	return &ReportController{DB: db, Cfg: cfg}
}

func (h *ReportController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/reports")
	g.Use(middleware.JWT(h.Cfg, h.DB))
	g.GET("/students/:id/summary", h.studentSummary)
	g.GET("/courses/:id/summary", h.courseSummary)
	g.GET("/sessions/:id/summary", h.sessionSummary)
}

func (h *ReportController) studentSummary(c *gin.Context) {
	sid := c.Param("id")
	var gradeRows []struct {
		Scale     string
		ValueNum  *float64
		ValuePass *bool
		MaxPts    *float64
	}
	if err := h.DB.Table("assessment_grades").
		Select("assessment_grades.scale, assessment_grades.value_num, assessment_grades.value_pass, assessments.max_pts").
		Joins("JOIN assessments ON assessments.id = assessment_grades.assessment_id").
		Where("assessment_grades.student_id = ?", sid).
		Find(&gradeRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}

	var attendanceRows []struct {
		Status string
	}
	if err := h.DB.Table("attendances").
		Select("attendances.status").
		Where("attendances.student_id = ?", sid).
		Find(&attendanceRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}

    gradeInputs := make([]gradeInput, 0, len(gradeRows))
    for _, row := range gradeRows {
        gradeInputs = append(gradeInputs, gradeInput{Scale: row.Scale, ValueNum: row.ValueNum, ValuePass: row.ValuePass, MaxPts: row.MaxPts})
    }
    attendanceInputs := make([]attendanceInput, 0, len(attendanceRows))
    for _, row := range attendanceRows {
        attendanceInputs = append(attendanceInputs, attendanceInput{Status: row.Status})
    }

    stats := aggregateGrades(gradeInputs)
    attendance := aggregateAttendance(attendanceInputs)

    c.JSON(http.StatusOK, resp.StudentSummaryReport{
        StudentID:        sid,
        AverageGrade:     round2(stats.average),
        PassRate:         round4(stats.passRate),
        AttendanceRate:   round4(attendance.rate),
        TotalGrades:      stats.total,
        TotalAttendances: attendance.total,
    })
}

func (h *ReportController) courseSummary(c *gin.Context) {
	cid := c.Param("id")
	var gradeRows []struct {
		StudentID string
		GroupID   *string
		Scale     string
		ValueNum  *float64
		ValuePass *bool
		MaxPts    *float64
	}
	if err := h.DB.Table("assessment_grades").
		Select("assessment_grades.student_id, students.group_id, assessment_grades.scale, assessment_grades.value_num, assessment_grades.value_pass, assessments.max_pts").
		Joins("JOIN assessments ON assessments.id = assessment_grades.assessment_id").
		Joins("JOIN students ON students.id = assessment_grades.student_id").
		Where("assessments.course_id = ?", cid).
		Find(&gradeRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}

	var attendanceRows []struct {
		StudentID string
		Status    string
		GroupID   *string
	}
	if err := h.DB.Table("attendances").
		Select("attendances.student_id, attendances.status, students.group_id").
		Joins("JOIN lessons ON lessons.id = attendances.lesson_id").
		Joins("JOIN students ON students.id = attendances.student_id").
		Where("lessons.course_id = ?", cid).
		Find(&attendanceRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}

    gradeInputs := make([]gradeInput, 0, len(gradeRows))
    for _, row := range gradeRows {
        gradeInputs = append(gradeInputs, gradeInput{Scale: row.Scale, ValueNum: row.ValueNum, ValuePass: row.ValuePass, MaxPts: row.MaxPts})
    }
    attendanceInputs := make([]attendanceInput, 0, len(attendanceRows))
    for _, row := range attendanceRows {
        attendanceInputs = append(attendanceInputs, attendanceInput{Status: row.Status})
    }

    stats := aggregateGrades(gradeInputs)
    attendance := aggregateAttendance(attendanceInputs)

	studentIDs := make(map[string]struct{})
	groupIDs := make(map[string]struct{})
	for _, row := range gradeRows {
		if row.StudentID != "" {
			studentIDs[row.StudentID] = struct{}{}
		}
		if row.GroupID != nil && *row.GroupID != "" {
			groupIDs[*row.GroupID] = struct{}{}
		}
	}
	for _, row := range attendanceRows {
		if row.StudentID != "" {
			studentIDs[row.StudentID] = struct{}{}
		}
		if row.GroupID != nil && *row.GroupID != "" {
			groupIDs[*row.GroupID] = struct{}{}
		}
	}

	c.JSON(http.StatusOK, resp.CourseSummaryReport{
		CourseID:       cid,
		AverageGrade:   round2(stats.average),
		PassRate:       round4(stats.passRate),
		AttendanceRate: round4(attendance.rate),
		TotalGrades:    stats.total,
		StudentCount:   len(studentIDs),
		UniqueGroups:   len(groupIDs),
	})
}

func (h *ReportController) sessionSummary(c *gin.Context) {
	sid := c.Param("id")
	var gradeRows []struct {
		StudentID string
		CourseID  string
		Scale     string
		ValueNum  *float64
		ValuePass *bool
		MaxPts    *float64
	}
	if err := h.DB.Table("assessment_grades").
		Select("assessment_grades.student_id, assessments.course_id, assessment_grades.scale, assessment_grades.value_num, assessment_grades.value_pass, assessments.max_pts").
		Joins("JOIN assessments ON assessments.id = assessment_grades.assessment_id").
		Joins("JOIN courses ON courses.id = assessments.course_id").
		Where("courses.academic_session_id = ?", sid).
		Find(&gradeRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}

	var attendanceRows []struct {
		StudentID string
		Status    string
		CourseID  string
	}
	if err := h.DB.Table("attendances").
		Select("attendances.student_id, attendances.status, lessons.course_id").
		Joins("JOIN lessons ON lessons.id = attendances.lesson_id").
		Joins("JOIN courses ON courses.id = lessons.course_id").
		Where("courses.academic_session_id = ?", sid).
		Find(&attendanceRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}

    gradeInputs := make([]gradeInput, 0, len(gradeRows))
    for _, row := range gradeRows {
        gradeInputs = append(gradeInputs, gradeInput{Scale: row.Scale, ValueNum: row.ValueNum, ValuePass: row.ValuePass, MaxPts: row.MaxPts})
    }
    attendanceInputs := make([]attendanceInput, 0, len(attendanceRows))
    for _, row := range attendanceRows {
        attendanceInputs = append(attendanceInputs, attendanceInput{Status: row.Status})
    }

    stats := aggregateGrades(gradeInputs)
    attendance := aggregateAttendance(attendanceInputs)

	studentIDs := make(map[string]struct{})
	courseIDs := make(map[string]struct{})
	for _, row := range gradeRows {
		if row.StudentID != "" {
			studentIDs[row.StudentID] = struct{}{}
		}
		if row.CourseID != "" {
			courseIDs[row.CourseID] = struct{}{}
		}
	}
	for _, row := range attendanceRows {
		if row.StudentID != "" {
			studentIDs[row.StudentID] = struct{}{}
		}
		if row.CourseID != "" {
			courseIDs[row.CourseID] = struct{}{}
		}
	}

	var courseCount int64
	_ = h.DB.Model(&m.Course{}).Where("academic_session_id = ?", sid).Count(&courseCount)

	c.JSON(http.StatusOK, resp.SessionSummaryReport{
		SessionID:      sid,
		AverageGrade:   round2(stats.average),
		PassRate:       round4(stats.passRate),
		AttendanceRate: round4(attendance.rate),
		TotalGrades:    stats.total,
		StudentCount:   len(studentIDs),
		CourseCount:    int(courseCount),
	})
}
