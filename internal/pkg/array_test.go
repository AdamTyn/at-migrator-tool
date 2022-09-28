package pkg

import (
	"testing"
)

func TestArrayStr_Has(t *testing.T) {
	var testArr ArrayString = []string{"a", "b", "unless"}
	__print := func(ins ...string) {
		for k := range ins {
			if !testArr.Has(ins[k]) {
				t.Logf("%s is not in testArr", ins[k])
			}
		}
	}
	__print("110", "s", "a")
}

func TestArrayInt_Has(t *testing.T) {
	var testArr ArrayInt = []int{1, 2, 3, 4}
	__print := func(ins ...int) {
		for k := range ins {
			if !testArr.Has(ins[k]) {
				t.Logf("%d is not in testArr", ins[k])
			}
		}
	}
	__print(1, 10)
}
