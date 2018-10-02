package db

//KeyPair rsa's public & private key pair
type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

type retentionStats struct {
	date       int64
	register   int64 //注册人数
	loginDay1  int64 //1,2,3,7,14,30登录
	loginDay2  int64
	loginDay3  int64
	loginDay7  int64
	loginDay14 int64
	loginDay30 int64
}
