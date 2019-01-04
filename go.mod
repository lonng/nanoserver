module github.com/lonng/nanoserver

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20181001203147-e3636079e1a4
	golang.org/x/net => github.com/golang/net v0.0.0-20180926154720-4dfa2610cdf3
	golang.org/x/sys => github.com/golang/sys v0.0.0-20180928133829-e4b3c5e90611
	golang.org/x/text => github.com/golang/text v0.3.0
)

require (
	cloud.google.com/go v0.28.0 // indirect
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/chanxuehong/rand v0.0.0-20180830053958-4b3aff17f488 // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20180901172138-1eb28afdf9b6 // indirect
	github.com/go-sql-driver/mysql v1.4.0
	github.com/go-xorm/core v0.6.0
	github.com/go-xorm/xorm v0.7.0
	github.com/google/go-cmp v0.2.0 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.2
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/lib/pq v1.0.0 // indirect
	github.com/lonng/nano v0.4.0
	github.com/lonng/nex v1.4.1
	github.com/mattn/go-sqlite3 v1.9.0 // indirect
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.8.0
	github.com/sirupsen/logrus v1.1.0
	github.com/spf13/viper v1.2.1
	github.com/urfave/cli v1.20.0
	github.com/xxtea/xxtea-go v0.0.0-20170828040851-35c4b17eecf6
	github.com/ziutek/mymysql v1.5.4 // indirect
	golang.org/x/crypto v0.0.0-20180904163835-0709b304e793
	golang.org/x/net v0.0.0-20180926154720-4dfa2610cdf3
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f // indirect
	golang.org/x/text v0.3.0
	google.golang.org/appengine v1.2.0 // indirect
	gopkg.in/chanxuehong/wechat.v2 v2.0.0-20180924084534-7e0579cb5377
)
