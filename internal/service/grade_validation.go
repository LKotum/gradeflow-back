package service

var allowedGradeValues = map[float32]struct{}{
	2: {},
	3: {},
	4: {},
	5: {},
}

func isGradeValueAllowed(value float32) bool {
	_, ok := allowedGradeValues[value]
	return ok
}
