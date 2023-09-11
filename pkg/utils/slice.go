package utils

import (
	"strings"
)

func SliceFilter[T any](slice []T, f func(T) bool) (result []T) {
	for _, v := range slice {
		if f(v) {
			result = append(result, v)
		}
	}

	return
}

func SliceStringFilter(d []string) []string {
	return SliceFilter(d, func(s string) bool {
		return strings.Trim(s, " ") != ""
	})
}

func SliceUnique[T comparable](d *[]T) {
	t := 0
	f := make(map[T]bool, len(*d))

	for i, v := range *d {
		if _, ok := f[v]; !ok {
			f[v] = true
			(*d)[t] = (*d)[i]
			t++
		}
	}

	*d = (*d)[:t]
}

func SliceIn[T comparable](item T, d []T) bool {
	for _, v := range d {
		if v == item {
			return true
		}
	}

	return false
}

func SliceColumn[T any](slice []map[string]T, key string) (result []T) {
	for _, m := range slice {
		result = append(result, m[key])
	}
	return
}

func SliceStringDiff(in1, in2 []string) (out1, out2 []string) {
	for _, h := range in1 {
		if SliceIn(h, in2) {
			continue
		}

		out1 = append(out1, h)
	}

	for _, h := range in2 {
		if SliceIn(h, in1) {
			continue
		}

		out2 = append(out2, h)
	}

	return
}
