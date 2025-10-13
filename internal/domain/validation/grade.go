package validation

import (
	"fmt"
	"math"
)

const (
	GradeScalePoints   = "points"
	GradeScalePassFail = "passfail"
	GradeScaleFive     = "five"
	GradeScaleHundred  = "hundred"
)

var gradeScales = map[string]struct{}{
	GradeScalePoints:   {},
	GradeScalePassFail: {},
	GradeScaleFive:     {},
	GradeScaleHundred:  {},
}

var assessmentTypeScaleRules = map[string]map[string]struct{}{
	"zachet": {
		GradeScalePassFail: {},
	},
	"diff_zachet": {
		GradeScaleFive: {},
	},
	"exam": {
		GradeScaleFive:    {},
		GradeScaleHundred: {},
	},
}

// ValidateGradeScale ensures that provided scale is recognised by the system.
func ValidateGradeScale(scale string) error {
	if _, ok := gradeScales[scale]; !ok {
		return fmt.Errorf("invalid scale: %s", scale)
	}
	return nil
}

// ValidateAssessmentTypeScale ensures that assessment type (if recognised) matches allowed scale set.
// Unknown types are only checked for valid scale value.
func ValidateAssessmentTypeScale(assessmentType string, scale string) error {
	if err := ValidateGradeScale(scale); err != nil {
		return err
	}
	if allowed, ok := assessmentTypeScaleRules[assessmentType]; ok {
		if _, ok2 := allowed[scale]; !ok2 {
			return fmt.Errorf("scale %s is not allowed for assessment type %s", scale, assessmentType)
		}
	}
	return nil
}

// ValidateGradeValues verifies that grade values correspond to the scale semantics.
func ValidateGradeValues(scale string, valueNum *float64, valuePass *bool) error {
	switch scale {
	case GradeScalePassFail:
		if valuePass == nil {
			return fmt.Errorf("valuePass required for %s scale", scale)
		}
		if valueNum != nil {
			return fmt.Errorf("valueNum must be empty for %s scale", scale)
		}
	case GradeScalePoints:
		if valueNum == nil {
			return fmt.Errorf("valueNum required for %s scale", scale)
		}
		if valuePass != nil {
			return fmt.Errorf("valuePass must be empty for %s scale", scale)
		}
	case GradeScaleFive:
		if valueNum == nil {
			return fmt.Errorf("valueNum required for %s scale", scale)
		}
		if valuePass != nil {
			return fmt.Errorf("valuePass must be empty for %s scale", scale)
		}
		v := *valueNum
		if math.Trunc(v) != v {
			return fmt.Errorf("valueNum must be an integer 2..5 for %s scale", scale)
		}
		if v < 2 || v > 5 {
			return fmt.Errorf("valueNum must be between 2 and 5 for %s scale", scale)
		}
	case GradeScaleHundred:
		if valueNum == nil {
			return fmt.Errorf("valueNum required for %s scale", scale)
		}
		if valuePass != nil {
			return fmt.Errorf("valuePass must be empty for %s scale", scale)
		}
		if *valueNum < 0 || *valueNum > 100 {
			return fmt.Errorf("valueNum must be between 0 and 100 for %s scale", scale)
		}
	default:
		return fmt.Errorf("invalid scale: %s", scale)
	}
	return nil
}
