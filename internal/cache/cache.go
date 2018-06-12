//Package cache implement a cache which backend by redis
package cache

import (
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/go-kit/kit/log"

	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/types"
)

const (
	defaultConnTimeout  = 5 * time.Second
	defaultReadTimeout  = 1 * time.Second
	defaultWriteTimeout = 5 * time.Second
)

const (
	cmdSet       = "SET"
	cmdSetExpire = "SETEX"
	cmdGet       = "GET"
	cmdExists    = "EXISTS"
	cmdDel       = "DEL"
	cmdExpire    = "EXPIRE"
	cmdTTL       = "TTL"
	cmdSelect    = "SELECT"
	cmdGetSet    = "GETSET"
	cmdHMSet     = "HMSET"
	cmdHGetAll   = "HGETALL"
	cmdIncr      = "INCR"
)

//Cache cache the token and others
type cache struct {
	logger log.Logger
	pool   *redis.Pool
}

var c *cache
var once sync.Once

// Synonyms of Redis SET command
func Set(k string, v interface{}) error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := conn.Do(cmdSet, k, v)
	if err != nil {
		c.logger.Log("error", err)
		err = errutil.YXErrCacheOperation
	}
	return err
}

// Synonyms of Redis SETEX command
// save a key-value pair with expired time( by second)
func SetExpire(k string, v interface{}, expire uint32) error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := conn.Do(cmdSetExpire, k, expire, v)
	if err != nil {
		c.logger.Log("error", err)
		err = errutil.YXErrCacheOperation
		return err
	}

	return nil
}

// Synonyms of Redis EXPIRE command
// Expire set a key-value pair's expired time( by second)
func Expire(k string, expire int) error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := conn.Do(cmdExpire, k, expire)
	if err != nil {
		c.logger.Log("error", err)
		err = errutil.YXErrCacheOperation
	}
	return err
}

// Synonyms of Redis GET command
// Get get the value by key
func Get(key string) (interface{}, error) {
	conn := c.pool.Get()
	defer conn.Close()

	ret, err := conn.Do(cmdGet, key)
	if err != nil {
		c.logger.Log("error", err)
		return nil, errutil.YXErrCacheOperation
	}
	return ret, err
}

// Synonyms of Redis DEL command
// Delete delete a key-value
func Delete(key string) error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := conn.Do(cmdDel, key)
	if err != nil {
		c.logger.Log("error", err)
		err = errutil.YXErrCacheOperation
	}
	return err

}

// Synonyms of Redis TTL command
// TTL the ttl of the key
func TTL(key string) (int, error) {
	conn := c.pool.Get()
	defer conn.Close()

	ttl, err := redis.Int(conn.Do(cmdTTL, key))
	if err != nil {
		c.logger.Log("error", err)
		err = errutil.YXErrCacheOperation
		return 0, err
	}
	return ttl, nil
}

// Update sets key to value
func Update(key string, value interface{}) error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := conn.Do(cmdGetSet, key, value)
	if err != nil {
		c.logger.Log("error", err)
		err = errutil.YXErrCacheOperation
	}
	return err
}

// Synonyms of Redis EXISTS command
// Check the whether key-value existed.
func Exists(key string) bool {
	conn := c.pool.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do(cmdExists, key))
	if err != nil {
		c.logger.Log("error", err)
	}
	return ok
}

func String(key string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do(cmdExists, key))
	if err != nil {
		c.logger.Log("error", err)
		return "", err
	}

	if !ok {
		c.logger.Log("err", "cache not found: "+key)
		return "", errutil.YXErrNotFound
	}

	s, err := redis.String(conn.Do(cmdGet, key))
	if err != nil {
		c.logger.Log("error", err)
		return "", err
	}
	return s, nil
}

func Int(key string) (int, error) {
	conn := c.pool.Get()
	defer conn.Close()

	i, err := redis.Int(conn.Do(cmdGet, key))
	if err != nil {
		c.logger.Log("error", err.Error())
		return 0, err
	}
	return i, err
}

func SetStruct(k string, v interface{}) error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := conn.Do(cmdHMSet, redis.Args{k}.AddFlat(v)...)
	if err != nil {
		c.logger.Log("error", err)
		err = errutil.YXErrCacheOperation
	}
	return err
}

func SetStructExpire(k string, v interface{}, expire int) error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := conn.Do(cmdHMSet, redis.Args{k}.AddFlat(v)...)
	if err != nil {
		c.logger.Log("error", err)
		err = errutil.YXErrCacheOperation
	}

	_, err = conn.Do(cmdExpire, k, expire)
	if err != nil {
		c.logger.Log("error", err)
		err = errutil.YXErrCacheOperation
	}
	return err
}

func Struct(token string, v interface{}) error {
	conn := c.pool.Get()
	defer conn.Close()

	values, err := redis.Values(conn.Do(cmdHGetAll, token))
	if err != nil {
		c.logger.Log("error", err)
		return errutil.YXErrCacheOperation
	}

	if err := redis.ScanStruct(values, v); err != nil {
		c.logger.Log("error", err)
		return errutil.YXErrCacheOperation
	}
	return nil
}

func IncrKey(key string) (int, error) {
	conn := c.pool.Get()
	defer conn.Close()

	i, err := redis.Int(conn.Do(cmdIncr, key))
	if err != nil {
		return 0, err
	}
	return i, nil
}

//MustBootUp boot up a shard cache, it will panic if failed
func MustBootUp(logger log.Logger, addr string) types.Closer {
	if logger == nil || addr == "" {
		panic(errutil.YXErrIllegalParameter)
	}

	var closer types.Closer

	f := func() {
		logger = log.NewContext(logger).With("commponent", "cache")

		pool := &redis.Pool{
			MaxIdle:     32,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", addr,
					redis.DialConnectTimeout(defaultConnTimeout),
					redis.DialReadTimeout(defaultReadTimeout),
					redis.DialWriteTimeout(defaultWriteTimeout))

				if err != nil {
					return nil, err
				}
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}
				_, err := c.Do("PING")
				return err
			},
		}

		c = &cache{
			pool:   pool,
			logger: logger,
		}

		logger.Log("msg", "running")
		closer = func() {
			pool.Close()
			logger.Log("msg", "stopped")
		}
	}

	once.Do(f)
	return closer
}
