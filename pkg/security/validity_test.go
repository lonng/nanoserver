package security

import "testing"

var testdata = []string{
	"123123kj123",
	"sdflsjsdfsdf",
	"SDdko300df",
}

var testdata2 = []string{
	"skdfd",
	" sdkfjf",
	"23409.sdf0#",
	"!sdlkfj,/ lksjf",
}

func TestValidateName(t *testing.T) {
	for _, name := range testdata {
		if !ValidateName(name) {
			t.Errorf("should pass name: %s", name)
			t.Fail()
		}
	}

	for _, name := range testdata2 {
		if ValidateName(name) {
			t.Errorf("should not pass name: %s", name)
			t.Fail()
		}
	}
}
