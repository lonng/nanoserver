//Package algoutil contain some scaffold algo
package algoutil

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	mathrand "math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/lonng/nanoserver/pkg/errutil"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	//"triple/modules/security"
	"github.com/lonng/nanoserver/pkg/crypto"
)

//MD5String md5 digest in string
func MD5String(plain string) string {
	cipher := MD5([]byte(plain))
	return hex.EncodeToString(cipher)
}

//MD5 md5 digest
func MD5(plain []byte) []byte {
	md5Ctx := md5.New()
	md5Ctx.Write(plain)
	cipher := md5Ctx.Sum(nil)
	return cipher[:]
}

//CallSite the caller's file & line
func CallSite() interface{} {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	}
	return string(file + ":" + strconv.FormatInt(int64(line), 10))
}

//GenRSAKey gen a rsa key pair, the bit size is 512
func GenRSAKey() (privateKey, publicKey string, err error) {
	//public gen the private key
	privKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		return "", "", err
	}

	derStream := x509.MarshalPKCS1PrivateKey(privKey)
	privateKey = base64.StdEncoding.EncodeToString(derStream)

	//gen the public key
	pubKey := &privKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", "", err
	}
	publicKey = base64.StdEncoding.EncodeToString(derPkix)
	return privateKey, publicKey, nil
}

// RSAEncrypt encrypt data by rsa
func RSAEncrypt(plain []byte, pubKey string) ([]byte, error) {
	buf, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}
	p, err := x509.ParsePKIXPublicKey(buf)
	if err != nil {
		return nil, err
	}
	if pub, ok := p.(*rsa.PublicKey); ok {
		return rsa.EncryptPKCS1v15(rand.Reader, pub, plain) //RSA算法加密
	}
	return nil, errutil.ErrIllegalParameter
}

// RsaDecrypt decrypt data by rsa
func RSADecrypt(cipher []byte, privKey string) ([]byte, error) {
	if cipher == nil {
		return nil, errutil.ErrIllegalParameter
	}
	buf, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		return nil, err
	}
	priv, err := x509.ParsePKCS1PrivateKey(buf)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, cipher) //RSA解密算法
}

// Sign with database appsecret string(base64 encode)
func Sign(plain []byte, privKey string) (string, error) {
	if plain == nil {
		return "", errutil.ErrIllegalParameter
	}
	buf, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		return "", err
	}
	priv, err := x509.ParsePKCS1PrivateKey(buf)
	if err != nil {
		return "", err
	}
	return crypto.Sign(priv, plain)
}

// Verify with database appkey string(base64 encode)
func Verify(pubKey string, data []byte, sign string) error {

	buf, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return err
	}
	p, err := x509.ParsePKIXPublicKey(buf)
	if err != nil {
		return err
	}
	if pub, ok := p.(*rsa.PublicKey); ok {
		return crypto.Verify(pub, data, sign)
	}
	return errutil.ErrIllegalParameter
}

func VerifyRSAWithMD5(pubKey string, data []byte, sign string) error {
	buf, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return err
	}
	p, err := x509.ParsePKIXPublicKey(buf)
	if err != nil {
		return err
	}
	if pub, ok := p.(*rsa.PublicKey); ok {
		return crypto.VerifyRSAWithMD5(pub, data, sign)
	}
	return errutil.ErrIllegalParameter
}

//func MaskPhone(phone string) (string, error) {
//	if !security.ValidatePhone(phone) {
//		return "", errutil.ErrWrongPhoneNumber
//	}
//	return fmt.Sprintf("%s****%s", phone[:3], phone[7:]), nil
//}

// 生成随机字符串
func RandStr(strlen int) string {
	mathrand.Seed(time.Now().Unix())
	data := make([]byte, strlen)
	var num int
	for i := 0; i < strlen; i++ {
		num = mathrand.Intn(57) + 65
		for {
			if num > 90 && num < 97 {
				num = mathrand.Intn(57) + 65
			} else {
				break
			}
		}
		data[i] = byte(num)
	}
	return string(data)
}

func Utf8ToGBK(utf8str string) string {
	result, _, _ := transform.String(simplifiedchinese.GBK.NewEncoder(), utf8str)
	return result
}

func AccessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		h.ServeHTTP(w, r)
	})
}

func OptionControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			json.NewEncoder(w).Encode(`{ "code": 0, "data": "success"}`)
			return
		}

		h.ServeHTTP(w, r)
	})
}

//HTTPGet http's get method
func HTTPGet(url string) (string, error) {
	var body []byte
	rspn, err := http.Get(url)

	if err != nil {
		return "", err
	}
	defer rspn.Body.Close()
	body, err = ioutil.ReadAll(rspn.Body)

	return string(body), err
}

func CopyFile(dst, src string) error {
	if dst == "" || src == "" {
		return errutil.ErrIllegalParameter
	}

	srcDir, _ := filepath.Split(src)
	// get properties of source dir
	srcDirInfo, err := os.Stat(srcDir)
	if err != nil {
		return err
	}

	dstDir, _ := filepath.Split(dst)
	if err != nil {
		return err
	}

	MakeDirIfNeed(dstDir, srcDirInfo.Mode())

	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	return err
}

func CopyDir(dst string, src string) error {
	// get properties of source dir
	srcDirInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// create dest dir
	err = MakeDirIfNeed(dst, srcDirInfo.Mode())
	if err != nil {
		return err
	}

	srcDir, _ := os.Open(src)
	objs, err := srcDir.Readdir(-1)
	if err != nil {
		return err
	}

	const sep = string(filepath.Separator)
	for _, obj := range objs {
		srcFile := src + sep + obj.Name()
		dstFile := dst + sep + obj.Name()

		if obj.IsDir() {
			// create sub-directories - recursively
			if err = CopyDir(dstFile, srcFile); err != nil {
				return err
			}
			continue
		}

		err = CopyFile(dstFile, srcFile)
		if err != nil {
			return err
		}

	}
	return err
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func MakeDirIfNeed(dir string, mode os.FileMode) error {
	dir = strings.TrimRight(dir, "/")

	if FileExists(dir) {
		return nil
	}

	err := os.MkdirAll(dir, mode)
	return err
}

func Unused(args ...interface{}) {}

func RunCmd(cmdName string, workingDir string, args ...string) (string, error) {
	const duration = time.Second * 7200

	cmd := exec.Command(cmdName, args...)

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}

	chanErr := make(chan error)
	go func() {
		multiReader := io.MultiReader(stdout, stderr)
		in := bufio.NewScanner(multiReader)
		for in.Scan() {
			buf.Write(in.Bytes())
			buf.WriteString("\n")
		}

		if err := in.Err(); err != nil {
			chanErr <- err
			return
		}

		close(chanErr)

	}()

	// wait or timeout
	chanDone := make(chan error)

	go func() {
		chanDone <- cmd.Wait()
	}()
	select {
	case <-time.After(duration):
		cmd.Process.Kill()
		return "", fmt.Errorf("run command: %s failed with timeout", cmdName)

	case err, ok := <-chanErr:
		if ok {
			return "", err
		}

	case e := <-chanDone:
		fmt.Printf("error %+v\n", e)
	}

	return buf.String(), nil
}

//TimeRange adjust the time range.
func TimeRange(start, end int64) (int64, int64) {
	if start < 0 {
		start = 0
	}
	if end < 0 || end > time.Now().Unix() {
		end = time.Now().Unix()
	}

	if start > end {
		start, end = end, start
	}

	return start, end
}
