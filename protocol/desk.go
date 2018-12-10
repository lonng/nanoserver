package protocol

import (
	"github.com/lonng/nanoserver/pkg/constant"
)

type EnterDeskInfo struct {
	DeskPos  int    `json:"deskPos"`
	Uid      int64  `json:"acId"`
	Nickname string `json:"nickname"`
	IsReady  bool   `json:"isReady"`
	Sex      int    `json:"sex"`
	IsExit   bool   `json:"isExit"`
	HeadUrl  string `json:"headURL"`
	Score    int    `json:"score"`
	IP       string `json:"ip"`
	Offline  bool   `json:"offline"`
}

type ExitResponse struct {
	AccountId int64 `json:"acid"`
	IsExit    bool  `json:"isexit"`
	ExitType  int   `json:"exitType"`
	DeskPos   int   `json:"deskPos"`
}

type PlayerEnterDesk struct {
	Data []EnterDeskInfo `json:"data"`
}

type ExitRequest struct {
	IsDestroy bool `json:"isDestroy"`
}

type QueItem struct {
	Uid int64 `json:"uid"`
	Que int   `json:"que"`
}

type DingQue struct {
	Que int `json:"que"`
}

type DeskBasicInfo struct {
	DeskID string `json:"deskId"`
	Title  string `json:"title"`
	Desc   string `json:"desc"`
	Mode   int    `json:"mode"`
}

type ScoreInfo struct {
	Uid   int64 `json:"acId"`
	Score int   `json:"score"`
}

type DuanPaiInfo struct {
	Uid    int64 `json:"acId"`
	OnHand []int `json:"mjs"`
}

type DuanPai struct {
	MarkerID    int64         `json:"markerId"` //庄家
	Dice1       int           `json:"dice1"`
	Dice2       int           `json:"dice2"`
	AccountInfo []DuanPaiInfo `json:"accountInfo"`
}

type Op struct {
	Type    int   `json:"op"`
	TileIDs []int `json:"mjidxs"` //
}

type OpTypeDo struct {
	Uid     []int64 `json:"uid"`
	OpType  int     `json:"optype"`
	HuType  int     `json:"hutype"`
	TileIDs []int   `json:"mjs"`
}

type MoPai struct {
	AccountID int64 `json:"acId"`
	TileIDs   []int `json:"mjids"`
}

type GangPaiScoreChange struct {
	IsXiaYu bool        `json:"isXiaYu"`
	Changes []ScoreInfo `json:"changes"`
}

type BeHuInfo struct {
	AccountID int   `json:"acid"`
	FanShu    int   `json:"fanshu"`
	Coin      int64 `json:"coin"`
}

type RoundStats struct {
	FanNum     int    `json:"fanshu"`     //番数
	Feng       int    `json:"feng"`       //刮风
	Yu         int    `json:"yu"`         //下雨
	Total      int    `json:"total"`      //总分
	BannerType int    `json:"bannerType"` //显示[1]自摸,[2]胡,[3]点炮,[4]赔付图标
	Desc       string `json:"desc"`
}

type MatchStats struct {
	ZiMoNum     int    `json:"ziMo"`        //自摸次数
	HuNum       int    `json:"hu"`          //和数
	PaoNum      int    `json:"pao"`         //炮数
	AnGangNum   int    `json:"anGang"`      //暗杠数
	MingGangNum int    `json:"mingGang"`    //明杠数
	TotalScore  int    `json:"totalScore"`  //总分
	Uid         int64  `json:"uid"`         //id
	Account     string `json:"account"`     //名字
	IsPaoWang   bool   `json:"isPaoWang"`   //是否是炮王
	IsBigWinner bool   `json:"isBigWinner"` //是否是大赢家
	IsCreator   bool   `json:"isCreator"`   //是否是房主
}

type HuInfo struct {
	Uid           int64       `json:"acId"`
	HuPaiType     HuPaiType   `json:"huPaiType"`
	ScoreChange   []ScoreInfo `json:"scoreChange"`
	TotalWinScore int         `json:"totalWinScore"` //赢的所有输家的总分数
}

type HandTilesInfo struct {
	Uid    int64 `json:"acId"`
	Tiles  []int `json:"shouPai"`
	HuPai  int   `json:"huPai"`
	IsTing bool  `json:"isTing"`
}

type GameEndScoreChange struct {
	Uid    int64 `json:"acId"`
	Score  int   `json:"score"`
	Remain int   `json:"remain"`
}

type RoundOverStats struct {
	Title       string               `json:"title"`
	Round       string               `json:"round"`
	HandTiles   []*HandTilesInfo     `json:"tiles"`
	Stats       []*RoundStats        `json:"stats"`
	ScoreChange []GameEndScoreChange `json:"scoreChange"`
}

type DeskPlayerData struct {
	Uid        int64 `json:"acId"`
	HandTiles  []int `json:"shouPaiIds"`
	ChuTiles   []int `json:"chuPaiIds"`
	PGTiles    []int `json:"gangInfos"`
	LatestTile int   `json:"lastTile"`
	IsHu       bool  `json:"isHu"`
	HuPai      int   `json:"huPai"`
	HuType     int   `json:"huType"`
	Que        int   `json:"que"`
	Score      int   `json:"score"`
}

type SyncDesk struct {
	Status        constant.DeskStatus `json:"status"` //1,2,3,4,5
	Players       []DeskPlayerData    `json:"players"`
	ScoreInfo     []ScoreInfo         `json:"scoreInfo"`
	MarkerUid     int64               `json:"markerAcId"`
	LastMoPaiUid  int64               `json:"lastMoPaiAcId"`
	RestCount     int                 `json:"restCnt"`
	Dice1         int                 `json:"dice1"`
	Dice2         int                 `json:"dice2"`
	Hint          *Hint               `json:"hint"`
	LastTileId    int                 `json:"lastChuPaiId"`
	LastChuPaiUid int64               `json:"lastChuPaiUid"`
}

type DeskOptions struct {
	Mode     int `json:"mode"`
	MaxRound int `json:"round"`
	MaxFan   int `json:"maxFan"`

	Zimo string `json:"zimo"`

	// 玩法
	Menqing  bool `json:"menqing"`  // 门清中张
	Jiangdui bool `json:"jiangdui"` // 幺九将对
	Jiaxin   bool `json:"jiaxin"`   // 夹心五
	Pengpeng bool `json:"pengpeng"` // 碰碰胡两番
	Pinghu   bool `json:"pinghu"`   // 点炮可平胡
	Yaojiu   bool `json:"yaojiu"`   // 全幺九
}

type CreateDeskRequest struct {
	Version  string       `json:"version"` //客户端版本
	ClubId   int64        `json:"clubId"`  // 俱乐部ID
	DeskOpts *DeskOptions `json:"options"` // 游戏额外选项
}

type CreateDeskResponse struct {
	Code      int       `json:"code"`
	Error     string    `json:"error"`
	TableInfo TableInfo `json:"tableInfo"`
}

type ReConnect struct {
	Uid     int64  `json:"uid"`
	Name    string `json:"name"`
	HeadUrl string `json:"headUrl"`
	Sex     int    `json:"sex"`
}

type DeskListRequest struct {
	Player int64 `json:"player"`
	Offset int   `json:"offset"`
	Count  int   `json:"count"`
}

type Desk struct {
	Id           int64  `json:"id"`
	Creator      int64  `json:"creator"`
	Round        int    `json:"round"`
	DeskNo       string `json:"desk_no"`
	Mode         int    `json:"mode"`
	Player0      int64  `json:"player0"`
	Player1      int64  `json:"player1"`
	Player2      int64  `json:"player2"`
	Player3      int64  `json:"player3"`
	PlayerName0  string `json:"player_name0"`
	PlayerName1  string `json:"player_name1"`
	PlayerName2  string `json:"player_name2"`
	PlayerName3  string `json:"player_name3"`
	ScoreChange0 int    `json:"score_change0"`
	ScoreChange1 int    `json:"score_change1"`
	ScoreChange2 int    `json:"score_change2"`
	ScoreChange3 int    `json:"score_change3"`
	CreatedAt    int64  `json:"created_at"`
	CreatedAtStr string `json:"created_at_str"`
	DismissAt    int64  `json:"dismiss_at"`
	Extras       string `json:"extras"`
}

type DestroyDeskResponse struct {
	RoundStats       *RoundOverStats `json:"roundStats"`
	MatchStats       []MatchStats    `json:"stats"`
	Title            string          `json:"title"`
	IsNormalFinished bool            `json:"isNormalFinished"`
}

type DeskListResponse struct {
	Code  int    `json:"code"`
	Total int64  `json:"total"` //总数量
	Data  []Desk `json:"data"`
}

type DeleteDeskByIDRequest struct {
	ID string `json:"id"` //房间ID
}
type DeskByIDRequest struct {
	ID int64 `json:"id"` //房间ID
}

type DeskByIDResponse struct {
	Code int   `json:"code"`
	Data *Desk `json:"data"`
}

type LiangDaoHintMessage struct {
	AccountId int64 `json:"acid"`
	Tings     Tings `json:"tings"`
}

type LiangDaoMessage struct {
	AccountId int64 `json:"acid"`
	HuIndexs  []int `json:"hu"`
	KouIndexs []int `json:"kou"`
}

type RoundReady struct {
	Multiple int `json:"multiple"`
}

type JiangMa struct {
	ID  int `json:"id"`
	Fan int `json:"fan"`
}

type UnCompleteDeskResponse struct {
	Exist     bool      `json:"exist"`
	TableInfo TableInfo `json:"tableInfo"`
}

type RecordingVoice struct {
	FileId string `json:"fileId"`
}

type PlayRecordingVoice struct {
	Uid    int64  `json:"uid"`
	FileId string `json:"fileId"`
}

type ClientInitCompletedRequest struct {
	IsReEnter bool `json:"isReenter"`
}
