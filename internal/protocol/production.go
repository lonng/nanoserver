package protocol

type CreateProductionRequest struct {
	Name      string `json:"name"`       //名字
	Extra     string `json:"extra"`      //额外信息
	Currency  string `json:"currency"`   //币种
	Price     int    `json:"price"`      //金额(以分计）
	RealPrice int    `json:"real_price"` //实际充值金额
	Type      int    `json:"type"`       //类型, 1-SDK充值, 2-渠道
}

type UpdateProductionRequest struct {
	ProductionID string `json:"production_id"` //商品ID
	Name         string `json:"name"`          //名字
	Extra        string `json:"extra"`         //额外信息
	Currency     string `json:"currency"`      //币种
	Price        int    `json:"price"`         //金额
	RealPrice    int    `json:"real_price"`    //实际充值金额
	Type         int    `json:"type"`          //类型, 1-SDK充值, 2-渠道, 3-全部
}

type DeleteProductionRequest struct {
	ProductionID string `json:"production_id"` //商品ID
}

type ProductionListRequest struct {
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Type   int `json:"platform"` //产品类型: 1-贪玩蛇SDK产品, 2-其他渠道产品
}

//提供给SDK用的商品信息
type ProductionInfoLite struct {
	ProductionID string `json:"production_id"` //商品ID
	Name         string `json:"name"`          //名字
	Price        int    `json:"price"`         //金额
	RealPrice    int    `json:"real_price"`    //实际充值金额
}

//提供给后台管理工具用的商品信息
type ProductionInfo struct {
	ProductionInfoLite

	Currency string //币种
	OnlineAt int64  //上架时间
	//OfflineAt int64	   //下架时间
	Type  int    //产品类型: 代币还是游戏商品
	Extra string //额外信息
}

type ProductionListLiteResponse struct {
	Code int                  `json:"code"`
	Data []ProductionInfoLite `json:"data"`
}

type ProductionListResponse struct {
	Code  int              `json:"code"`
	Total int64            `json:"total"` //总数量
	Data  []ProductionInfo `json:"data"`
}
