package main

import (
	"slices"
	"testing"
)

func TestMapSlice(t *testing.T) {
	testCases := []struct {
		input []int32
		want  []int64
	}{
		{input: nil, want: nil},
		{input: []int32{}, want: []int64{}},
		{input: []int32{1}, want: []int64{2}},
		{input: []int32{1, 2}, want: []int64{2, 4}},
	}
	for _, c := range testCases {
		got := MapSlice(c.input, func(i int32) int64 {
			return 2 * int64(i)
		})
		if want := c.want; !slices.Equal(got, want) {
			t.Errorf("result mismatch, input=%v, got=%v, want=%v", c.input, got, want)
		}
	}
}

func TestSliceConcatDeDup(t *testing.T) {
	testCases := []struct {
		input [][]int
		want  []int
	}{
		{input: nil, want: nil},
		{input: [][]int{}, want: []int{}},
		{input: [][]int{{}, nil}, want: []int{}},
		{input: [][]int{{1}, {1, 2}, {2, 3}}, want: []int{1, 2, 3}},
	}
	for _, c := range testCases {
		if got, want := SliceConcatDeDup(c.input...), c.want; !slices.Equal(got, want) {
			t.Errorf("result mismatch, input=%v, got=%v, want=%v", c.input, got, want)
		}
	}
}
