package hashmap

import (
	"fmt"
	"testing"
	_ "unsafe"
)

type i1 interface {
}

type a1 struct {
	i1
}

func TestName(t *testing.T) {
	var aa interface{} = &a1{}
	switch aa.(type) {
	case i1:
		fmt.Println(aa)
	case a1:
		fmt.Println(aa,"a1")

	}
}
