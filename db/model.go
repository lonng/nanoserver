package db

import (
	"time"

	"github.com/lonng/nanoserver/db/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	log "github.com/sirupsen/logrus"
)

const asyncTaskBacklog = 128

var (
	database *xorm.Engine
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

				if _, err := database.Insert(t); err != nil {
					logger.Error(err)
				}

			case t, ok := <-chUpdate:
				if !ok {
					return
				}

				if _, err := database.Update(t); err != nil {
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
				database.Ping()
			}
		}
	}()
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
		database = db
	}

	// 设置日志相关
	database.SetLogger(&Logger{Entry: logger.WithField("orm", "xorm")})

	chWrite = make(chan interface{}, asyncTaskBacklog)
	chUpdate = make(chan interface{}, asyncTaskBacklog)

	// options
	database.SetMaxIdleConns(settings.maxIdleConns)
	database.SetMaxOpenConns(settings.maxOpenConns)
	database.ShowSQL(settings.showSQL)

	syncSchema()
	envInit()

	closer := func() {
		close(chWrite)
		close(chUpdate)
		database.Close()
		logger.Info("stopped")
	}

	return closer
}

func syncSchema() {
	database.StoreEngine("InnoDB").Sync2(
		new(model.Agent),
		new(model.CardConsume),
		new(model.Desk),
		new(model.History),
		new(model.Login),
		new(model.Online),
		new(model.Order),
		new(model.Recharge),
		new(model.Register),
		new(model.ThirdAccount),
		new(model.Trade),
		new(model.User),
		new(model.Uuid),
		new(model.Club),
		new(model.UserClub),
	)
}
