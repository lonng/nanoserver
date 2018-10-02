package room

import "testing"

func TestNext(t *testing.T) {
	for i := 0; i < 10000; i++ {
		Next()
		//t.Log(Next())
	}
}
