package util

import (
	"math/rand"
	"time"
)

var mathR *rand.Rand

func init() {
	mathR = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func Int63n(n int64) int64 {
	return mathR.Int63n(n)
}

func SliceHas[T comparable](sl []T, v T) bool {
	for _, s := range sl {
		if v == s {
			return true
		}
	}
	return false
}

func Ptr[T any](t T) *T {
	return &t
}

func RemoveDuplicates[T comparable](sl []T, excludes ...T) []T {
	res := make([]T, 0, len(sl)/2)
	for _, s := range sl {
		if !SliceHas(res, s) && !SliceHas(excludes, s) {
			res = append(res, s)
		}
	}
	return res
}
