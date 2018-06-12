package provider

import (
	"strings"
	"sync"

	"github.com/go-kit/kit/log"

	"github.com/lonnng/nanoserver/db/model"
	"github.com/lonnng/nanoserver/internal/errutil"
)

type Provider interface {
	Setup(log.Logger) error
	CreateOrderResponse(*model.Order) (interface{}, error)
	Notify(request interface{}) (trade *model.Trade, response interface{}, err error)
}

type providerMgr struct {
	sync.RWMutex
	providers map[string]Provider
	logger    log.Logger
}

var pm = &providerMgr{providers: make(map[string]Provider)}

func (pm *providerMgr) setup(logger log.Logger) error {
	pm.logger = logger
	//所有的provider, 配置启动参数
	for name, p := range pm.providers {
		logger.Log("msg", "register provider: "+name)
		if err := p.Setup(logger); err != nil {
			return err
		}
	}
	return nil
}

func (pm *providerMgr) register(p string, handler Provider) error {
	//p = strings.TrimSpace(p)
	if p == "" || handler == nil {
		return errutil.YXErrInvalidParameter
	}

	pm.Lock()
	defer pm.Unlock()

	//existed
	if _, ok := pm.providers[p]; ok {
		return nil
	}

	pm.providers[p] = handler
	return nil
}

func (pm *providerMgr) provider(p string) Provider {
	p = strings.TrimSpace(p)
	if p == "" {
		return nil
	}

	pm.RLock()
	defer pm.RUnlock()

	if p, ok := pm.providers[p]; ok {
		return p
	}

	return nil
}

func Register(p string, handler Provider) error {
	return pm.register(p, handler)
}

func WithName(p string) Provider {
	return pm.provider(p)
}

func MustSetup(logger log.Logger) {
	if logger == nil {
		panic(errutil.YXErrInvalidParameter)
	}

	logger = log.NewContext(logger).With("component", "provider")
	if err := pm.setup(logger); err != nil {
		panic(err)
	}
	logger.Log("msg", "running")
}
