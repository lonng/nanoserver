package provider

import (
	"github.com/lonnng/nanoserver/db/model"
)

type Base struct{}

func (b *Base) CreateOrderResponse(*model.Order) (interface{}, error) {
	panic("CreateOrder must be overrideed")
}

func (b *Base) Notify(request interface{}) (*model.Trade, interface{}, error) {
	panic("PayNotify must be overrideed")
}
