package main

import (
	"golang.org/x/exp/maps"
)

func MapSlice[E1, E2 any](x []E1, conv func(e E1) E2) []E2 {
	var res []E2
	for _, e := range x {
		res = append(res, conv(e))
	}
	return res
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

func SliceConcatDeDup[T comparable](x ...[]T) []T {
	if len(x) == 0 {
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
	return maps.Keys(s)
}
