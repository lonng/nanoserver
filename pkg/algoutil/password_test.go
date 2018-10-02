package algoutil

import "testing"

func TestPasswordHash(t *testing.T) {
	pwds := []string{
		"hsdlfjsdlfjsldkjfsldj",
		"你好!!!",
		"superadmin",
	}

	for _, pwd := range pwds {
		hash, salt := PasswordHash(pwd)
		t.Logf("hash: %s\nsalt: %s\n pwd: %s\n", hash, salt, pwd)
		if !VerifyPassword(pwd, salt, hash) {
			t.Fail()
		}
	}
}
