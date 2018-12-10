package protocol

import (
	"github.com/lonng/nanoserver/pkg/constant"
)

type GetRankInfoRequest struct {
	IsSelf bool `json:"isself"`
	Start  int  `json:"start"`
	Len    int  `json:"len"`
}

type MailOperateRequest struct {
	MailIDs []int64 `json:"mailids"`
}

type ApplyForDailyMatchRequest struct {
	Arg1           int `json:"arg1"`
	DailyMatchType int `json:"dailyMatchType"`
	Multiple       int `json:"multiple"`
}

type JQToCoinRequest struct {
	Count int `json:"count"`
}

type BashiCoinOpRequest struct {
	Op   int `json:"op"`
	Coin int `json:"coin"`
}

type ReJoinDeskRequest struct {
	DeskNo string `json:"deskId"`
}

type ReJoinDeskResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type ReEnterDeskRequest struct {
	DeskNo string `json:"deskId"`
}

type ReEnterDeskResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}
type JoinDeskRequest struct {
	Version string `json:"version"`
	//AccountId int64         `json:"acId"`
	DeskNo string `json:"deskId"`
}

type TableInfo struct {
	DeskNo    string              `json:"deskId"`
	CreatedAt int64               `json:"createdAt"`
	Creator   int64               `json:"creator"`
	Title     string              `json:"title"`
	Desc      string              `json:"desc"`
	Status    constant.DeskStatus `json:"status"`
	Round     uint32              `json:"round"`
	Mode      int                 `json:"mode"`
}

type JoinDeskResponse struct {
	Code      int       `json:"code"`
	Error     string    `json:"error"`
	TableInfo TableInfo `json:"tableInfo"`
}

type DestoryDeskRequest struct {
	DeskNo string `json:"deskId"`
}

//选择执行的动作
type OpChoosed struct {
	Type   int
	TileID int
}

type MingAction struct {
	KouIndexs []int `json:"kou"` //index
	ChuPaiID  int   `json:"chu"`
	HuIndexs  []int `json:"hu"`
}

type OpChooseRequest struct {
	OpType int `json:"optype"`
	Index  int `json:"idx"`
}

type ChooseOneScoreRequest struct {
	Pos int `json:"pos"`
}

type CheckOrderReqeust struct {
	OrderID string `json:"orderid"`
}

type CheckOrderResponse struct {
	Code   int    `json:"code"`
	Error  string `json:"error"`
	FangKa int    `json:"fangka"`
}
type FanPaiRequest struct {
	Pos        int  `json:"pos"`
	IsUseCoin  bool `json:"isUseCoin"`
	IsMultiple bool `json:"isMultiple"`
	OpType     int  `json:"opType"`
}

type DissolveStatusItem struct {
	DeskPos int    `json:"deskPos"`
	Status  string `json:"status"`
}

type DissolveResponse struct {
	DissolveUid    int64                `json:"dissolveUid"`
	DissolveStatus []DissolveStatusItem `json:"dissolveStatus"`
	RestTime       int32                `json:"restTime"`
}

type DissolveStatusRequest struct {
	Result bool `json:"result"`
}

type DissolveStatusResponse struct {
	DissolveStatus []DissolveStatusItem `json:"dissolveStatus"`
	RestTime       int32                `json:"restTime"`
}

type DissolveResult struct {
	DeskPos int `json:"deskPos"`
}

type CalcLastTingRequest struct {
	KouIndexs  []int `json:"kou"`
	TingIndexs []int `json:"ting"`
}

type CalcLastTingResponse struct {
	ForbidIndexs []int `json:"forbid"`
	Tings        Tings `json:"tings"`
}

type PlayerOfflineStatus struct {
	Uid     int64 `json:"uid"`
	Offline bool  `json:"offline"`
}

type CoinChangeInformation struct {
	Coin int64 `json:"coin"`
}
