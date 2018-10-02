package algoutil

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

func pemParse(data []byte, pemType string) ([]byte, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("No PEM block found")
	}
	if pemType != "" && block.Type != pemType {
		return nil, fmt.Errorf("Key's type is '%s', expected '%s'", block.Type, pemType)
	}
	return block.Bytes, nil
}

func ParsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	pemData, err := pemParse(data, "RSA PRIVATE KEY")
	if err != nil {
		return nil, err
	}

	return x509.ParsePKCS1PrivateKey(pemData)
}

func LoadPrivateKey(privKeyPath string) (*rsa.PrivateKey, error) {
	certPEMBlock, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return nil, err
	}

	return ParsePrivateKey(certPEMBlock)
}
