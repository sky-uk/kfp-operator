package apis

func SliceDiff[T any](as, bs []T, cmp func(T, T) bool) []T {
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

func Find[T any](ts []T, predicate func(T) bool) (*T, bool) {
	for _, t := range ts {
		if predicate(t) {
			return &t, true
		}
	}
	return nil, false
}

func MapErr[R, S any](rs []R, mapFn func(R) (S, error)) (ss []S, err error) {
	ss = make([]S, len(rs))

	for i, r := range rs {
		ss[i], err = mapFn(r)
		if err != nil {
			return
		}
	}

	return
}
