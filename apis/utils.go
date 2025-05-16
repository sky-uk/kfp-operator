package apis

import (
	"github.com/samber/lo"
	"maps"
)

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

func Forall[T any](ts []T, predicate func(T) bool) bool {
	for _, t := range ts {
		if !predicate(t) {
			return false
		}
	}

	return true
}

func Find[T any](ts []T, predicate func(T) bool) (*T, bool) {
	for _, t := range ts {
		if predicate(t) {
			return &t, true
		}
	}
	return nil, false
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

func FlatMapErr[R, S any](rs []R, mapFn func(R) ([]S, error)) ([]S, error) {
	ss, err := MapErr(rs, mapFn)
	if err != nil {
		return nil, err
	}

	return lo.Flatten(ss), nil
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

func Duplicates[R comparable](in []R) (out []R) {
	duplicatesInOutput := make(map[R]bool)

	for _, i := range in {
		inOutput, duplicate := duplicatesInOutput[i]
		if inOutput {
			continue
		}

		if duplicate {
			out = append(out, i)
			duplicatesInOutput[i] = true
		} else {
			duplicatesInOutput[i] = false
		}
	}

	return
}

func MapValues[K comparable, V, W any](vs map[K]V, mapValueFn func(V) W) map[K]W {
	kws := make(map[K]W)

	for name, value := range vs {
		kws[name] = mapValueFn(value)
	}

	return kws
}

func MapConcat[K comparable, V any](m1, m2 map[K]V) map[K]V {
	m3 := make(map[K]V)

	maps.Copy(m3, m1)
	maps.Copy(m3, m2)

	return m3
}
