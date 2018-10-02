package algoutil

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pborman/uuid"
)

func passwordHash(pwd, salt string) string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "%s%x%s", salt, pwd, salt)
	result1 := sha1.Sum(buf.Bytes())

	buf.Reset()

	fmt.Fprintf(buf, "%s%s%x%s%s", pwd, salt, result1, salt, pwd)
	result2 := sha1.Sum(buf.Bytes())

	buf.Reset()
	fmt.Fprintf(buf, "%x", result2)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// PasswordHash accept password and generate with uuid as salt
// FORMAT: sha1.Sum(pwd + salt + sha1.Sum(salt + pwd + salt) + salt + pwd)
func PasswordHash(pwd string) (hash, salt string) {
	salt = strings.Replace(uuid.New(), "-", "", -1)
	hash = passwordHash(pwd, salt)
	return hash, salt
}

func VerifyPassword(pwd, salt, hash string) bool {
	return passwordHash(pwd, salt) == hash
}
