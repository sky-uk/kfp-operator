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

func Filter[T any](ts []T, filter func(T) bool) (filtered []T) {
	for _, t := range ts {
		if filter(t) {
			filtered = append(filtered, t)
		}
	}

	return
}

func Map[R, S any](rs []R, mapFn func(R) S) []S {
	ss := make([]S, len(rs))

	for i, r := range rs {
		ss[i] = mapFn(r)
	}

	return ss
}

func ToMap[R any, K comparable, V any](rs []R, mapFn func(R) (K, V)) map[K]V {
	ss := make(map[K]V, len(rs))

	for _, r := range rs {
		k, v := mapFn(r)
		ss[k] = v
	}

	return ss
}

func Collect[R, S any](rs []R, mapFn func(R) (S, bool)) []S {
	var ss []S

	for _, r := range rs {
		mapped, canMap := mapFn(r)
		if canMap {
			ss = append(ss, mapped)
		}
	}

	return ss
}