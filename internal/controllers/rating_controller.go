package controllers

import (
    "net/http"
    "sort"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "gradeflow/internal/config"
    req "gradeflow/internal/domain/dto/request"
    resp "gradeflow/internal/domain/dto/response"
    m "gradeflow/internal/domain/models"
    "gradeflow/pkg/middleware"
)

const (
    defaultRatingLimit = 20
    maxRatingLimit     = 100
)

type RatingController struct {
    DB  *gorm.DB
    Cfg config.Config
}

func NewRatingController(db *gorm.DB, cfg config.Config) *RatingController {
    return &RatingController{DB: db, Cfg: cfg}
}

func (h *RatingController) RegisterRoutes(rg *gin.RouterGroup) {
    g := rg.Group("/ratings")
    g.Use(middleware.JWT(h.Cfg, h.DB))
    g.GET("/groups", h.groupRatings)
    g.GET("/students", h.studentRatings)
    g.GET("/courses/:id/groups", h.courseGroupRatings)
    g.GET("/groups/:id/students", h.groupStudentRatings)
}

func (h *RatingController) groupRatings(c *gin.Context) {
    var q req.GroupRatingQuery
    _ = c.ShouldBindQuery(&q)
    limit := normalizeLimit(q.Limit)
    items, total, err := h.computeGroupRatings(q.CourseID, q.SessionID, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
        return
    }
    c.JSON(http.StatusOK, resp.List[resp.GroupRatingEntry]{
        Items: items,
        Page: resp.Page{Limit: limit, Offset: 0, Total: int64(total)},
    })
}

func (h *RatingController) courseGroupRatings(c *gin.Context) {
    courseID := c.Param("id")
    var q req.LimitQuery
    _ = c.ShouldBindQuery(&q)
    limit := normalizeLimit(q.Limit)
    items, total, err := h.computeGroupRatings(courseID, "", limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
        return
    }
    c.JSON(http.StatusOK, resp.List[resp.GroupRatingEntry]{
        Items: items,
        Page: resp.Page{Limit: limit, Offset: 0, Total: int64(total)},
    })
}

func (h *RatingController) groupStudentRatings(c *gin.Context) {
    groupID := c.Param("id")
    var q req.LimitQuery
    _ = c.ShouldBindQuery(&q)
    limit := normalizeLimit(q.Limit)
    items, total, err := h.computeStudentRatings("", "", groupID, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
        return
    }
    c.JSON(http.StatusOK, resp.List[resp.StudentRatingEntry]{
        Items: items,
        Page: resp.Page{Limit: limit, Offset: 0, Total: int64(total)},
    })
}

func (h *RatingController) studentRatings(c *gin.Context) {
    var q req.StudentRatingQuery
    _ = c.ShouldBindQuery(&q)
    limit := normalizeLimit(q.Limit)
    items, total, err := h.computeStudentRatings(q.CourseID, q.SessionID, q.GroupID, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
        return
    }
    c.JSON(http.StatusOK, resp.List[resp.StudentRatingEntry]{
        Items: items,
        Page: resp.Page{Limit: limit, Offset: 0, Total: int64(total)},
    })
}

func (h *RatingController) computeGroupRatings(courseID, sessionID string, limit int) ([]resp.GroupRatingEntry, int, error) {
    gradeQuery := h.DB.Table("assessment_grades ag").
        Select("ag.student_id, students.group_id, ag.scale, ag.value_num, ag.value_pass, a.max_pts").
        Joins("JOIN assessments a ON a.id = ag.assessment_id").
        Joins("JOIN students ON students.id = ag.student_id")
    if courseID != "" {
        gradeQuery = gradeQuery.Where("a.course_id = ?", courseID)
    }
    if sessionID != "" {
        gradeQuery = gradeQuery.Joins("JOIN courses course_filter ON course_filter.id = a.course_id").Where("course_filter.academic_session_id = ?", sessionID)
    }
    var gradeRows []struct {
        StudentID string
        GroupID   *string
        Scale     string
        ValueNum  *float64
        ValuePass *bool
        MaxPts    *float64
    }
    if err := gradeQuery.Find(&gradeRows).Error; err != nil {
        return nil, 0, err
    }

    attendanceQuery := h.DB.Table("attendances at").
        Select("at.student_id, students.group_id, at.status").
        Joins("JOIN lessons ON lessons.id = at.lesson_id").
        Joins("JOIN students ON students.id = at.student_id")
    if courseID != "" {
        attendanceQuery = attendanceQuery.Where("lessons.course_id = ?", courseID)
    }
    if sessionID != "" {
        attendanceQuery = attendanceQuery.Joins("JOIN courses course_att ON course_att.id = lessons.course_id").Where("course_att.academic_session_id = ?", sessionID)
    }
    var attendanceRows []struct {
        StudentID string
        GroupID   *string
        Status    string
    }
    if err := attendanceQuery.Find(&attendanceRows).Error; err != nil {
        return nil, 0, err
    }

    type groupAccumulator struct {
        grades     []gradeInput
        attendance []attendanceInput
        students   map[string]struct{}
    }

    accMap := make(map[string]*groupAccumulator)

    for _, row := range gradeRows {
        if row.GroupID == nil || *row.GroupID == "" {
            continue
        }
        gid := *row.GroupID
        acc := accMap[gid]
        if acc == nil {
            acc = &groupAccumulator{students: make(map[string]struct{})}
            accMap[gid] = acc
        }
        acc.grades = append(acc.grades, gradeInput{Scale: row.Scale, ValueNum: row.ValueNum, ValuePass: row.ValuePass, MaxPts: row.MaxPts})
        if row.StudentID != "" {
            acc.students[row.StudentID] = struct{}{}
        }
    }

    for _, row := range attendanceRows {
        if row.GroupID == nil || *row.GroupID == "" {
            continue
        }
        gid := *row.GroupID
        acc := accMap[gid]
        if acc == nil {
            acc = &groupAccumulator{students: make(map[string]struct{})}
            accMap[gid] = acc
        }
        acc.attendance = append(acc.attendance, attendanceInput{Status: row.Status})
        if row.StudentID != "" {
            acc.students[row.StudentID] = struct{}{}
        }
    }

    groupIDs := make([]string, 0, len(accMap))
    for gid := range accMap {
        groupIDs = append(groupIDs, gid)
    }
    nameMap, err := h.loadGroupNames(groupIDs)
    if err != nil {
        return nil, 0, err
    }

    type groupScore struct {
        id           string
        name         *string
        avg          float64
        passRate     float64
        attendance   float64
        studentCount int
    }

    scores := make([]groupScore, 0, len(accMap))
    for gid, acc := range accMap {
        if len(acc.grades) == 0 && len(acc.attendance) == 0 {
            continue
        }
        gradeStats := aggregateGrades(acc.grades)
        attendanceStats := aggregateAttendance(acc.attendance)
        scores = append(scores, groupScore{
            id:           gid,
            name:         nameMap[gid],
            avg:          gradeStats.average,
            passRate:     gradeStats.passRate,
            attendance:   attendanceStats.rate,
            studentCount: len(acc.students),
        })
    }

    total := len(scores)
    if total == 0 {
        return []resp.GroupRatingEntry{}, 0, nil
    }

    sort.Slice(scores, func(i, j int) bool {
        if scores[i].avg == scores[j].avg {
            if scores[i].passRate == scores[j].passRate {
                if scores[i].attendance == scores[j].attendance {
                    return scores[i].id < scores[j].id
                }
                return scores[i].attendance > scores[j].attendance
            }
            return scores[i].passRate > scores[j].passRate
        }
        return scores[i].avg > scores[j].avg
    })

    if limit > 0 && len(scores) > limit {
        scores = scores[:limit]
    }

    result := make([]resp.GroupRatingEntry, len(scores))
    for i, sc := range scores {
        result[i] = resp.GroupRatingEntry{
            GroupID:        sc.id,
            GroupName:      sc.name,
            AverageGrade:   round2(sc.avg),
            PassRate:       round4(sc.passRate),
            AttendanceRate: round4(sc.attendance),
            StudentCount:   sc.studentCount,
            Rank:           i + 1,
        }
    }

    return result, total, nil
}

func (h *RatingController) computeStudentRatings(courseID, sessionID, groupID string, limit int) ([]resp.StudentRatingEntry, int, error) {
    gradeQuery := h.DB.Table("assessment_grades ag").
        Select("ag.student_id, students.full_name, students.group_id, ag.scale, ag.value_num, ag.value_pass, a.max_pts").
        Joins("JOIN assessments a ON a.id = ag.assessment_id").
        Joins("JOIN students ON students.id = ag.student_id")
    if courseID != "" {
        gradeQuery = gradeQuery.Where("a.course_id = ?", courseID)
    }
    if sessionID != "" {
        gradeQuery = gradeQuery.Joins("JOIN courses course_filter ON course_filter.id = a.course_id").Where("course_filter.academic_session_id = ?", sessionID)
    }
    if groupID != "" {
        gradeQuery = gradeQuery.Where("students.group_id = ?", groupID)
    }
    var gradeRows []struct {
        StudentID string
        FullName  string
        GroupID   *string
        Scale     string
        ValueNum  *float64
        ValuePass *bool
        MaxPts    *float64
    }
    if err := gradeQuery.Find(&gradeRows).Error; err != nil {
        return nil, 0, err
    }

    attendanceQuery := h.DB.Table("attendances at").
        Select("at.student_id, students.full_name, students.group_id, at.status").
        Joins("JOIN lessons ON lessons.id = at.lesson_id").
        Joins("JOIN students ON students.id = at.student_id")
    if courseID != "" {
        attendanceQuery = attendanceQuery.Where("lessons.course_id = ?", courseID)
    }
    if sessionID != "" {
        attendanceQuery = attendanceQuery.Joins("JOIN courses course_att ON course_att.id = lessons.course_id").Where("course_att.academic_session_id = ?", sessionID)
    }
    if groupID != "" {
        attendanceQuery = attendanceQuery.Where("students.group_id = ?", groupID)
    }
    var attendanceRows []struct {
        StudentID string
        FullName  string
        GroupID   *string
        Status    string
    }
    if err := attendanceQuery.Find(&attendanceRows).Error; err != nil {
        return nil, 0, err
    }

    type studentAccumulator struct {
        name       string
        groupID    *string
        grades     []gradeInput
        attendance []attendanceInput
    }

    accMap := make(map[string]*studentAccumulator)

    for _, row := range gradeRows {
        acc := accMap[row.StudentID]
        if acc == nil {
            acc = &studentAccumulator{}
            accMap[row.StudentID] = acc
        }
        acc.name = row.FullName
        if row.GroupID != nil && *row.GroupID != "" {
            gid := *row.GroupID
            acc.groupID = &gid
        }
        acc.grades = append(acc.grades, gradeInput{Scale: row.Scale, ValueNum: row.ValueNum, ValuePass: row.ValuePass, MaxPts: row.MaxPts})
    }

    for _, row := range attendanceRows {
        acc := accMap[row.StudentID]
        if acc == nil {
            acc = &studentAccumulator{}
            accMap[row.StudentID] = acc
        }
        if acc.name == "" {
            acc.name = row.FullName
        }
        if acc.groupID == nil && row.GroupID != nil && *row.GroupID != "" {
            gid := *row.GroupID
            acc.groupID = &gid
        }
        acc.attendance = append(acc.attendance, attendanceInput{Status: row.Status})
    }

    type studentScore struct {
        id         string
        name       string
        groupID    *string
        avg        float64
        passRate   float64
        attendance float64
    }

    scores := make([]studentScore, 0, len(accMap))
    for sid, acc := range accMap {
        if len(acc.grades) == 0 && len(acc.attendance) == 0 {
            continue
        }
        gradeStats := aggregateGrades(acc.grades)
        attendanceStats := aggregateAttendance(acc.attendance)
        scores = append(scores, studentScore{
            id:         sid,
            name:       acc.name,
            groupID:    acc.groupID,
            avg:        gradeStats.average,
            passRate:   gradeStats.passRate,
            attendance: attendanceStats.rate,
        })
    }

    total := len(scores)
    if total == 0 {
        return []resp.StudentRatingEntry{}, 0, nil
    }

    sort.Slice(scores, func(i, j int) bool {
        if scores[i].avg == scores[j].avg {
            if scores[i].passRate == scores[j].passRate {
                if scores[i].attendance == scores[j].attendance {
                    return scores[i].id < scores[j].id
                }
                return scores[i].attendance > scores[j].attendance
            }
            return scores[i].passRate > scores[j].passRate
        }
        return scores[i].avg > scores[j].avg
    })

    if limit > 0 && len(scores) > limit {
        scores = scores[:limit]
    }

    result := make([]resp.StudentRatingEntry, len(scores))
    for i, sc := range scores {
        result[i] = resp.StudentRatingEntry{
            StudentID:      sc.id,
            FullName:       sc.name,
            GroupID:        copyStringPtr(sc.groupID),
            AverageGrade:   round2(sc.avg),
            PassRate:       round4(sc.passRate),
            AttendanceRate: round4(sc.attendance),
            Rank:           i + 1,
        }
    }

    return result, total, nil
}

func (h *RatingController) loadGroupNames(ids []string) (map[string]*string, error) {
    result := make(map[string]*string, len(ids))
    if len(ids) == 0 {
        return result, nil
    }
    var groups []struct {
        ID   string
        Name string
    }
    if err := h.DB.Model(&m.Group{}).Select("id, name").Where("id IN ?", ids).Find(&groups).Error; err != nil {
        return nil, err
    }
    for _, g := range groups {
        name := g.Name
        result[g.ID] = &name
    }
    return result, nil
}

func normalizeLimit(l int) int {
    if l <= 0 {
        return defaultRatingLimit
    }
    if l > maxRatingLimit {
        return maxRatingLimit
    }
    return l
}

func copyStringPtr(src *string) *string {
    if src == nil {
        return nil
    }
    val := *src
    return &val
}
