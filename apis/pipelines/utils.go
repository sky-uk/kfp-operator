package pipelines

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

func GroupMap[R any, K comparable, V any](rs []R, groupFn func(R) (K, V)) map[K][]V {
	vvs := make(map[K][]V)

	for _, r := range rs {
		k, v := groupFn(r)
		vvs[k] = append(vvs[k], v)
	}

	return vvs
}
