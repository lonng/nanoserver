module github.com/lonng/nanoserver

go 1.14

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20181001203147-e3636079e1a4
	golang.org/x/net => github.com/golang/net v0.0.0-20180926154720-4dfa2610cdf3
	golang.org/x/sys => github.com/golang/sys v0.0.0-20180928133829-e4b3c5e90611
	golang.org/x/text => github.com/golang/text v0.3.0
)

require (
	github.com/chanxuehong/rand v0.0.0-20180830053958-4b3aff17f488 // indirect
	github.com/go-sql-driver/mysql v1.4.0
	github.com/go-xorm/core v0.6.0
	github.com/go-xorm/xorm v0.7.0
	github.com/gorilla/mux v1.6.2
	github.com/lonng/nano v0.4.1-0.20190704005402-15209d995681
	github.com/lonng/nex v1.4.1
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.8.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/viper v1.2.1
	github.com/urfave/cli v1.20.1-0.20190203184040-693af58b4d51
	github.com/xxtea/xxtea-go v0.0.0-20170828040851-35c4b17eecf6
	golang.org/x/crypto v0.0.0-20190510104115-cbcb75029529
	golang.org/x/net v0.0.0-20190509222800-a4d6f7feada5
	golang.org/x/text v0.3.2
	gopkg.in/chanxuehong/wechat.v2 v2.0.0-20180924084534-7e0579cb5377
)
