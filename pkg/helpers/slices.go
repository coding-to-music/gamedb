package helpers

import (
	"math"
	"strconv"
	"strings"

	"github.com/gamedb/gamedb/pkg/log"
)

func StringSliceToIntSlice(in []string) (ret []int) {

	for _, v := range in {
		v = strings.TrimSpace(v)
		if v != "" {
			i, err := strconv.Atoi(v)
			log.Err(err)
			if err == nil {
				ret = append(ret, i)
			}
		}
	}
	return ret
}

func SliceHasInt(slice []int, i int) bool {
	for _, v := range slice {
		if v == i {
			return true
		}
	}
	return false
}

func SliceHasInt64(slice []int64, i int64) bool {
	for _, v := range slice {
		if v == i {
			return true
		}
	}
	return false
}

func SliceHasString(i string, slice []string) bool {
	for _, v := range slice {
		if v == i {
			return true
		}
	}
	return false
}

// todo, keep order by doing https://codereview.stackexchange.com/questions/191238/return-unique-items-in-a-go-slice
func UniqueInt(arg []int) []int {

	tempMap := make(map[int]uint8)

	for idx := range arg {
		tempMap[arg[idx]] = 0
	}

	tempSlice := make([]int, 0)
	for key := range tempMap {
		tempSlice = append(tempSlice, key)
	}
	return tempSlice
}

func UniqueInt64(arg []int64) []int64 {

	tempMap := make(map[int64]uint8)

	for idx := range arg {
		tempMap[arg[idx]] = 0
	}

	tempSlice := make([]int64, 0)
	for key := range tempMap {
		tempSlice = append(tempSlice, key)
	}
	return tempSlice
}

func FirstInts(slice []int, x int) []int {

	x = int(math.Min(float64(x), float64(len(slice))))
	return slice[0:x]
}

func Filter(ss []string, filter func(string) bool) (ret []string) {
	for _, s := range ss {
		if filter(s) {
			ret = append(ret, s)
		}
	}
	return
}
