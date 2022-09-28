package pkg

import "sort"

type ArrayString []string

func (arr ArrayString) Has(target string) bool {
	sort.Strings(arr)
	index := sort.SearchStrings(arr, target)
	if index < len(arr) && arr[index] == target {
		return true
	}
	return false
}

func (arr ArrayString) HasNot(target string) bool {
	return !arr.Has(target)
}

type ArrayInt []int

func (arr ArrayInt) Has(target int) bool {
	sort.Ints(arr)
	index := sort.SearchInts(arr, target)
	if index < len(arr) && arr[index] == target {
		return true
	}
	return false
}

func (arr ArrayInt) HasNot(target int) bool {
	return !arr.Has(target)
}
