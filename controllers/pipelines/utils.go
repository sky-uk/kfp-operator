package pipelines

func sliceDiff[T any](as, bs []T, cmp func(T, T) bool) []T {
	var diff []T
	for _, a := range as {
		exists := false
		for _, b := range bs {
			if cmp(a, b) {
				exists = true
			}
		}

		if !exists {
			diff = append(diff, a)
		}
	}

	return diff
}

func filter[T any](ts []T, filter func(T) bool) (filtered []T) {
	for _, t := range ts {
		if filter(t) {
			filtered = append(filtered, t)
		}
	}

	return
}
