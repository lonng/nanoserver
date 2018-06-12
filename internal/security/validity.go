package security

import (
	"regexp"
)

var (
	phoneRE *regexp.Regexp
	nameRE  *regexp.Regexp
)

func init() {
	r, err := regexp.Compile("^1[0-9]{10}$")
	if err != nil {
		panic(err.Error())
	}
	phoneRE = r

	r, err = regexp.Compile("^[0-9a-zA-Z.@]{6,32}$")
	if err != nil {
		panic(err.Error())
	}
	nameRE = r
}

func ValidateName(name string) bool {
	return nameRE.MatchString(name)
}

// 验证电话号码
func ValidatePhone(phone string) bool {
	return phoneRE.MatchString(phone)
}
