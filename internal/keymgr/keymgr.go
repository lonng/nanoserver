//Package keymgr manage all app's appkey & appsecret
package keymgr

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/lonnng/nanoserver/internal/algoutil"
	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/db"
)

//KeyPair a rsa key pair
type keyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

//KeyMgr all key pair's manager
type KeyMgr struct {
	logger log.Logger
	kps    map[string]*keyPair
	mu     sync.Mutex
}

func (km *KeyMgr) init(logger log.Logger) bool {
	km.kps = make(map[string]*keyPair)
	km.logger = log.NewContext(logger).With("commponent", "keymgr")
	return true
}

func (km *KeyMgr) loadKeyPairsHelper(pairs map[string]*db.KeyPair) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	for k, v := range pairs {
		pair, err := km.loadKeyPair(v)
		if err != nil {
			km.logger.Log("appid", k, "error", err)
			return err
		}

		km.kps[k] = pair
	}
	return nil

}

//loadKeyPair load a rsa key pair
func (km *KeyMgr) loadKeyPair(pair *db.KeyPair) (*keyPair, error) {
	pubKey, err := km.publicKey(pair.PublicKey)
	if err != nil {
		return nil, err
	}
	privKey, err := km.privateKey(pair.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &keyPair{
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}, nil
}

//privateKey load rsa private key
func (km *KeyMgr) privateKey(raw string) (*rsa.PrivateKey, error) {
	buf, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	rsaPriv, err := x509.ParsePKCS1PrivateKey(buf)
	if err != nil {
		return nil, err
	}
	return rsaPriv, nil
}

//privateKey load rsa public key
func (km *KeyMgr) publicKey(raw string) (*rsa.PublicKey, error) {
	buf, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	pubKey, err := x509.ParsePKIXPublicKey(buf)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, err
	}
	return rsaPub, nil
}

//LoadKeyPairs load all key pairs from database
func (km *KeyMgr) LoadKeyPairs() error {
	pairs, err := db.AppKeyPairs()
	if err != nil {
		km.logger.Log("error", err)
		return err
	}
	return km.loadKeyPairsHelper(pairs)
}

//New new a key pair manager
func New(logger log.Logger) *KeyMgr {
	if logger == nil {
		return nil
	}
	km := &KeyMgr{}
	if !km.init(logger) {
		return nil
	}
	return km
}

func (km *KeyMgr) isAppExisted(appID string) bool {
	km.mu.Lock()
	defer km.mu.Unlock()
	if _, ok := km.kps[appID]; ok {
		return true
	}
	return false
}

//KeyPairForApp the key pair belong to the app
func (km *KeyMgr) KeyPairForApp(appID string) *keyPair {
	if pair, ok := km.kps[appID]; ok {
		return pair
	}
	pair, err := db.KeyPairForApp(appID)
	if err != nil {
		km.logger.Log("error", err)
		return nil
	}

	kp, err := km.loadKeyPair(pair)
	if err != nil {
		km.logger.Log("error", err)
		return nil
	}

	km.mu.Lock()
	defer km.mu.Unlock()
	km.kps[appID] = kp
	return kp
}

//RegenKeyPairForApp  regenerate key pair for app
func (km *KeyMgr) RegenKeyPairForApp(appID string) error {
	if !km.isAppExisted(appID) {
		return errutil.YXErrNotFound
	}

	privKey, pubKey, err := algoutil.GenRSAKey()
	if err != nil {
		km.logger.Log("error", err)
		return errutil.YXErrServerInternal
	}

	app := &db.App{
		AppKey:    string(pubKey),
		AppSecret: string(privKey),
		Appid:     appID,
	}

	err = db.UpdateApp(app)
	if err != nil {
		km.logger.Log("error", err)
		return err
	}

	pair, err := km.loadKeyPair(
		&db.KeyPair{
			PublicKey:  pubKey,
			PrivateKey: privKey})
	if err != nil {
		km.logger.Log("error", err)
		return errutil.YXErrServerInternal
	}

	km.mu.Lock()
	defer km.mu.Unlock()
	km.kps[appID] = pair
	return nil
}
