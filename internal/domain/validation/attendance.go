package validation

import (
	"fmt"
	"sort"
	"strings"
)

const (
	AttendanceStatusPresent = "present"
	AttendanceStatusAbsent  = "absent"
	AttendanceStatusLate    = "late"
)

var attendanceStatuses = map[string]struct{}{
	AttendanceStatusPresent: {},
	AttendanceStatusAbsent:  {},
	AttendanceStatusLate:    {},
}

// ValidAttendanceStatuses returns a sorted copy of allowed attendance statuses.
func ValidAttendanceStatuses() []string {
	items := make([]string, 0, len(attendanceStatuses))
	for k := range attendanceStatuses {
		items = append(items, k)
	}
	sort.Strings(items)
	return items
}

func ValidateAttendanceStatus(status string) error {
	if _, ok := attendanceStatuses[status]; !ok {
		return fmt.Errorf("invalid status: %s (allowed: %s)", status, strings.Join(ValidAttendanceStatuses(), ", "))
	}
	return nil
}
