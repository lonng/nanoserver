package crypto

import (
	"crypto"
	"crypto/md5"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/lonng/nanoserver/pkg/errutil"
	"golang.org/x/crypto/pkcs12"
)

func ParsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	pemData, err := pemParse(data, "RSA PRIVATE KEY")
	if err != nil {
		return nil, err
	}

	return x509.ParsePKCS1PrivateKey(pemData)
}

func ParsePublicKey(data []byte) (*rsa.PublicKey, error) {
	pemData, err := pemParse(data, "PUBLIC KEY")
	if err != nil {
		return nil, err
	}

	keyInterface, err := x509.ParsePKIXPublicKey(pemData)
	if err != nil {
		return nil, err
	}

	pubKey, ok := keyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Could not cast parsed key to *rsa.PublickKey")
	}

	return pubKey, nil
}

func ParseCertSerialNo(data []byte) (string, error) {

	if data == nil {
		return "", errutil.ErrInvalidParameter
	}

	// Extract the PEM-encoded data block
	pemData, err := pemParse(data, "CERTIFICATE")
	if err != nil {
		return "", err
	}

	// Decode the certificate
	cert, err := x509.ParseCertificate(pemData)
	if err != nil {
		return "", fmt.Errorf("bad private key: %s", err)
	}

	return cert.SerialNumber.String(), nil
}

func LoadCertSerialNo(certPath string) (string, error) {

	pemData, err := ioutil.ReadFile(certPath)
	if err != nil {
		return "", err
	}
	return ParseCertSerialNo(pemData)
}

func pemParse(data []byte, pemType string) ([]byte, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("No PEM block found")
	}
	if pemType != "" && block.Type != pemType {
		return nil, fmt.Errorf("Key's type is '%s', expected '%s'", block.Type, pemType)
	}
	return block.Bytes, nil
}

func LoadPublicKey(pubKeyPath string) (*rsa.PublicKey, error) {
	certPEMBlock, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return nil, err
	}
	return ParsePublicKey(certPEMBlock)
}

func LoadPubKeyFromCert(certPath string) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	// Extract the PEM-encoded data block
	pemData, err := pemParse(data, "CERTIFICATE")
	if err != nil {
		return nil, err
	}

	// Decode the certificate
	cert, err := x509.ParseCertificate(pemData)
	if err != nil {
		return nil, fmt.Errorf("bad certificate : %s", err)
	}

	return cert.PublicKey.(*rsa.PublicKey), nil
}

func LoadPrivateKey(privKeyPath string) (*rsa.PrivateKey, error) {
	certPEMBlock, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return nil, err
	}

	return ParsePrivateKey(certPEMBlock)
}

func LoadPrivKeyAndCert(pfxPath string, password string) (*rsa.PrivateKey, *x509.Certificate, error) {

	pfxBlock, err := ioutil.ReadFile(pfxPath)
	if err != nil {
		return nil, nil, err
	}

	keyInterface, cert, err := pkcs12.Decode(pfxBlock, password)
	if err != nil {
		return nil, nil, err
	}
	return keyInterface.(*rsa.PrivateKey), cert, nil
}

func Sign(priKey *rsa.PrivateKey, data []byte) (string, error) {
	bs, err := rsa.SignPKCS1v15(nil, priKey, crypto.SHA1, SHA1Digest(data))
	if err != nil {
		return "", errutil.ErrSignFailed
	}

	return base64.StdEncoding.EncodeToString(bs), nil
}

func Verify(pubKey *rsa.PublicKey, data []byte, sign string) error {
	bs, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return errutil.ErrServerInternal
	}

	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA1, SHA1Digest(data), bs)
	if err != nil {
		return errutil.ErrVerifyFailed
	}
	return nil
}

func VerifyRSAWithMD5(pubKey *rsa.PublicKey, data []byte, sign string) error {
	bs, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return errutil.ErrServerInternal
	}

	err = rsa.VerifyPKCS1v15(pubKey, crypto.MD5, MD5Digest(data), bs)
	if err != nil {
		return errutil.ErrVerifyFailed
	}
	return nil
}

//SHA1Digest generate a digest
func SHA1Digest(data []byte) []byte {
	h := sha1.New()
	h.Write(data)

	return h.Sum(nil)
}

func MD5Digest(data []byte) []byte {
	h := md5.New()
	h.Write(data)

	return h.Sum(nil)
}
