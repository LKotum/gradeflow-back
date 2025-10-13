package controllers

import (
	"math"

	validation "gradeflow/internal/domain/validation"
)

type gradeInput struct {
	Scale     string
	ValueNum  *float64
	ValuePass *bool
	MaxPts    *float64
}

type gradeAggregation struct {
	total     int
	passCount int
	sum       float64
	average   float64
	passRate  float64
}

func aggregateGrades(rows []gradeInput) gradeAggregation {
	var agg gradeAggregation
	for _, row := range rows {
		if err := validation.ValidateGradeScale(row.Scale); err != nil {
			continue
		}
		score, passed, ok := normalizeGrade(row.Scale, row.ValueNum, row.ValuePass, row.MaxPts)
		if !ok {
			continue
		}
		agg.total++
		agg.sum += score
		if passed {
			agg.passCount++
		}
	}
	if agg.total > 0 {
		agg.average = agg.sum / float64(agg.total)
		agg.passRate = float64(agg.passCount) / float64(agg.total)
	}
	return agg
}

type attendanceInput struct {
	Status string
}

type attendanceAggregation struct {
	total int
	hits  int
	rate  float64
}

func aggregateAttendance(rows []attendanceInput) attendanceAggregation {
	var agg attendanceAggregation
	for _, row := range rows {
		agg.total++
		if row.Status == validation.AttendanceStatusPresent || row.Status == validation.AttendanceStatusLate {
			agg.hits++
		}
	}
	if agg.total > 0 {
		agg.rate = float64(agg.hits) / float64(agg.total)
	}
	return agg
}

func normalizeGrade(scale string, valueNum *float64, valuePass *bool, maxPts *float64) (float64, bool, bool) {
	switch scale {
	case validation.GradeScalePassFail:
		if valuePass == nil {
			return 0, false, false
		}
		if *valuePass {
			return 100, true, true
		}
		return 0, false, true
	case validation.GradeScaleHundred:
		if valueNum == nil {
			return 0, false, false
		}
		score := clamp(*valueNum, 0, 100)
		return score, score >= 60, true
	case validation.GradeScalePoints:
		if valueNum == nil {
			return 0, false, false
		}
		max := 100.0
		if maxPts != nil && *maxPts > 0 {
			max = *maxPts
		}
		if max <= 0 {
			max = 100
		}
		score := clamp((*valueNum/max)*100, 0, 100)
		return score, score >= 60, true
	case validation.GradeScaleFive:
		if valueNum == nil {
			return 0, false, false
		}
		score := clamp(((*valueNum-2)/3)*100, 0, 100)
		return score, *valueNum >= 3, true
	default:
		return 0, false, false
	}
}

func clamp(v float64, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
