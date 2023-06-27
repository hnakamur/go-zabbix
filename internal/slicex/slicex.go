package slicex

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func Map[E1, E2 any](x []E1, conv func(e E1) E2) []E2 {
	if x == nil {
		return nil
	}
	if len(x) == 0 {
		return []E2{}
	}
	res := make([]E2, len(x))
	for i, e := range x {
		res[i] = conv(e)
	}
	return res
}

func FailableMap[E1, E2 any](x []E1, conv func(e E1) (E2, error)) ([]E2, error) {
	if x == nil {
		return nil, nil
	}
	if len(x) == 0 {
		return []E2{}, nil
	}
	res := make([]E2, len(x))
	for i, e := range x {
		e2, err := conv(e)
		if err != nil {
			return nil, err
		}
		res[i] = e2
	}
	return res, nil
}

func ContainsDup[T comparable](x []T) bool {
	if len(x) == 0 {
		return false
	}

	s := make(map[T]struct{})
	for _, e := range x {
		if _, ok := s[e]; ok {
			return true
		}
		s[e] = struct{}{}
	}
	return false
}

func ConcatDeDup[T constraints.Ordered](x ...[]T) []T {
	if x == nil {
		return nil
	}
	s := make(map[T]struct{})
	for _, xx := range x {
		for _, e := range xx {
			if _, ok := s[e]; !ok {
				s[e] = struct{}{}
			}
		}
	}
	ret := maps.Keys(s)
	slices.Sort(ret)
	return ret
}
