package main

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func MapSlice[E1, E2 any](x []E1, conv func(e E1) E2) []E2 {
	var res []E2
	if x != nil {
		res = []E2{}
	}
	for _, e := range x {
		res = append(res, conv(e))
	}
	return res
}

func FailableMapSlice[E1, E2 any](x []E1, conv func(e E1) (E2, error)) ([]E2, error) {
	var res []E2
	if x != nil {
		res = []E2{}
	}
	for _, e := range x {
		e2, err := conv(e)
		if err != nil {
			return nil, err
		}
		res = append(res, e2)
	}
	return res, nil
}

func SliceContainsDup[T comparable](x []T) bool {
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

func SliceConcatDeDup[T constraints.Ordered](x ...[]T) []T {
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
