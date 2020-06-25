package memcache

import (
	"testing"

	"github.com/memcachier/mc"
)

type test struct {
	Val1 int
	Val2 string
}

func Test(t *testing.T) {

	client = mc.NewMC("localhost:11211", "", "")

	test1 := test{
		Val1: 3,
		Val2: "3",
	}

	// Set
	err := SetInterface("test", test1, 10)
	if err != nil {
		t.Error(err)
	}

	//
	test2 := test{}
	err = GetInterface("test", &test2)
	if err != nil {
		t.Error(err)
	}

	if test1.Val1 != test2.Val1 {
		t.Error("val1")
	}
	if test1.Val2 != test2.Val2 {
		t.Error("val2")
	}
}
