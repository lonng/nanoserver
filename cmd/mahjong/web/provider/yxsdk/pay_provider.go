package yxsdk

import (
	"strings"
	"sync"

	"github.com/lonnng/nanoserver/db/model"
	"github.com/lonnng/nanoserver/internal/errutil"

	"github.com/go-kit/kit/log"
)

type PayProvider interface {
	Setup() error
	CreateOrderResponse(*model.Order) (interface{}, error)
	Notify(request interface{}) (*model.Trade, interface{}, error)
}

type providerMgr struct {
	sync.RWMutex
	providers map[string]PayProvider
	logger    log.Logger
}

var pm = &providerMgr{providers: make(map[string]PayProvider)}

func (pm *providerMgr) setup(logger log.Logger) error {
	pm.logger = logger
	//所有的provider,配置启动参数
	for name, p := range pm.providers {
		logger.Log("msg", "register pay provider: "+name)
		if err := p.Setup(); err != nil {
			return err
		}
	}
	return nil
}

func (pm *providerMgr) register(p string, handler PayProvider) error {
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

func (pm *providerMgr) remove(p string) {
	pm.Lock()
	defer pm.Unlock()

	delete(pm.providers, p)
}

func (pm *providerMgr) provider(p string) PayProvider {
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

func Register(p string, handler PayProvider) error {
	return pm.register(p, handler)
}

func Remove(p string) {
	pm.remove(p)
}
