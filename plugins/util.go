package plugins

func IndexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}

	// Not found.
	return -1
}
