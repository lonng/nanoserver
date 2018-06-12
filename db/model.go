package db

import (
	"time"

	"github.com/lonnng/nanoserver/db/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	log "github.com/sirupsen/logrus"
)

const asyncTaskBacklog = 128

var (
	DB       *xorm.Engine
	logger   *log.Entry
	chWrite  chan interface{} // async write channel
	chUpdate chan interface{} // async update channel
)

type options struct {
	showSQL      bool
	maxOpenConns int
	maxIdleConns int
}

// ModelOption specifies an option for dialing a xordefaultModel.
type ModelOption func(*options)

// MaxIdleConns specifies the max idle connect numbers.
func MaxIdleConns(i int) ModelOption {
	return func(opts *options) {
		opts.maxIdleConns = i
	}
}

// MaxOpenConns specifies the max open connect numbers.
func MaxOpenConns(i int) ModelOption {
	return func(opts *options) {
		opts.maxOpenConns = i
	}
}

// ShowSQL specifies the buffer size.
func ShowSQL(show bool) ModelOption {
	return func(opts *options) {
		opts.showSQL = show
	}
}

func envInit() {
	// async task
	go func() {
		for {
			select {
			case t, ok := <-chWrite:
				if !ok {
					return
				}

				if _, err := DB.Insert(t); err != nil {
					logger.Error(err)
				}

			case t, ok := <-chUpdate:
				if !ok {
					return
				}

				if _, err := DB.Update(t); err != nil {
					logger.Error(err)
				}
			}
		}
	}()

	// 定时ping数据库, 保持连接池连接
	go func() {
		ticker := time.NewTicker(time.Minute * 5)
		for {
			select {
			case <-ticker.C:
				DB.Ping()
			}
		}
	}()

	if _, err := DB.Exec("alter table `user` auto_increment=1000000"); err != nil {
		panic(err)
	}

	//startupDBAutoCorrect()
}

//New create the database's connection
func MustStartup(dsn string, opts ...ModelOption) func() {
	logger = log.WithField("component", "model")
	settings := &options{
		maxIdleConns: defaultMaxConns,
		maxOpenConns: defaultMaxConns,
		showSQL:      true,
	}

	// options handle
	for _, opt := range opts {
		opt(settings)
	}

	logger.Infof("DSN=%s ShowSQL=%t MaxIdleConn=%v MaxOpenConn=%v", dsn, settings.showSQL, settings.maxIdleConns, settings.maxOpenConns)

	// create database instance
	if db, err := xorm.NewEngine("mysql", dsn); err != nil {
		panic(err)
	} else {
		DB = db
	}

	// 设置日志相关
	DB.SetLogger(&Logger{Entry: logger.WithField("orm", "xorm")})

	chWrite = make(chan interface{}, asyncTaskBacklog)
	chUpdate = make(chan interface{}, asyncTaskBacklog)

	// options
	DB.SetMaxIdleConns(settings.maxIdleConns)
	DB.SetMaxOpenConns(settings.maxOpenConns)
	DB.ShowSQL(settings.showSQL)

	syncSchema()
	envInit()

	closer := func() {
		close(chWrite)
		close(chUpdate)
		DB.Close()
		logger.Info("stopped")
	}

	return closer
}

func startupDBAutoCorrect() {
	// 定时对离线玩家的离线时间纠错
	//update `login` set logout_at = login.login_at + FLOOR(RAND() * 14400) + 600 where logout_at=0

	// 定时纠错玩家的离线状态
	//update `user` set is_online=1 where lastest_login_at < UNIX_TIMESTAMP() - 93600

	const rangeMin = 60 * 10          //10分钟
	const rangeInternal = 60 * 60 * 4 //4小时
	const timeOut = 60 * 60 * 24      //24小时超时

	//tick := time.NewTicker(60 * 60 * 4 * time.Second) //每4小时进行一次纠错
	unco := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			//case <-tick.C:
			//	kwx.logger.Log("msg", "database correct")
			//	_, err := DB.Exec(
			//		"UPDATE `login` SET logout_at = login.login_at + FLOOR(RAND() * ?) + ? WHERE logout_at=0",
			//		rangeInternal,
			//		rangeMin)
			//
			//	if err != nil {
			//		logger.Error( err)
			//	}
			//
			//	_, err = DB.Exec(
			//		"UPDATE `user` SET is_online=1 WHERE lastest_login_at < UNIX_TIMESTAMP() - ? AND is_online = 2",
			//		timeOut)
			//
			//	if err != nil {
			//		logger.Error( err)
			//	}
			case <-unco.C:
				// 修正username
				DB.Query("update rank set name=(select third_name from user left join third_account on user.id=third_account.uid where user.id=rank.uid);")
			}
		}
	}()
}

func syncSchema() {
	DB.StoreEngine("InnoDB").Sync2(
		new(model.AdminRecharge),
		new(model.Agent),
		new(model.App),
		new(model.CardConsume),
		new(model.Desk),
		new(model.History),
		new(model.Login),
		new(model.Online),
		new(model.Operation),
		new(model.Order),
		new(model.Production),
		new(model.Rank),
		new(model.Recharge),
		new(model.Register),
		new(model.ThirdAccount),
		new(model.ThirdProperty),
		new(model.Trade),
		new(model.User),
		new(model.Uuid),
		new(model.Club),
		new(model.UserClub),
	)
}
