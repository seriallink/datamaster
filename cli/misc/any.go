package misc

func NotIn(value any, list ...any) bool {
	for _, element := range list {
		if value == element {
			return false
		}
	}
	return true
}
