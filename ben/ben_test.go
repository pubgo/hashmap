package ben

import (
	"fmt"
	"testing"
)

type T struct {
	name string
}

func (t T) Name() string {
	return "Hi! " + t.name
}

func (t T) Name1() string {
	return "Hi! "
}

func TestName(t1 *testing.T) {
	t := T{name: "test"}
	fmt.Println(t.Name())  // Hi! test
	fmt.Println(T.Name(t)) // Hi! test
}
