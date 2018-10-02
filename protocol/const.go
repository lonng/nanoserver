package protocol

type HuPaiType int

//登录状态
const (
	LoginStatusSucc = 1
	LoginStatusFail = 2
)

const (
	ActionNewAccountSignIn    = "accountLogin"
	ActionGuestSignIn         = "anonymousLogin"
	ActionOldAccountSignIn    = "oldAccountLogin"
	ActionWebChatSignIn       = "webChatSignIn"
	ActionPhoneNumberRegister = "phoneRegister"
	ActionNormalRegister      = "normalRegister"
	ActionGetVerification     = "getVerification"
	ActionAccountRegister     = "accountRegister"
)

const (
	LoginTypeAuto   = "auto"
	LoginTypeManual = "manual"
)

const (
	VerificationTypeRegister = "register"
	VerificationTypeFindPW   = "findPW"
)

// 匹配类型
const (
	MatchTypeClassic = 1 //经典
	MatchTypeDaily   = 3 //每日匹配
)

const (
	CoinTypeSliver = 0 //银币
	CoinTypeGold   = 1 //金币
)

const (
	RoomTypeClassic      = 0
	RoomTypeDailyMatch   = 1
	RoomTypeMonthlyMatch = 2
	RoomTypeFinalMatch   = 3
)

const (
	DailyMatchLevelJunior = 0
	DailyMatchLevelSenior = 1
	DailyMatchLevelMaster = 2
)

const (
	ClassicLevelJunior = 0
	ClassicLevelMiddle = 1
	ClassicLevelSenior = 2
	ClassicLevelElite  = 3
	ClassicLevelMaster = 4
)

const (
	ExitTypeExitDeskUI           = -1
	ExitTypeDissolve             = 6
	ExitTypeSelfRequest          = 0
	ExitTypeClassicCoinNotEnough = 1
	ExitTypeDailyMatchEnd        = 2
	ExitTypeNotReadyForStart     = 3
	ExitTypeChangeDesk           = 4
	ExitTypeRepeatLogin          = 5
)

const (
	DeskStatusZb      = 1
	DeskStatusDq      = 2
	DeskStatusPlaying = 3
	DeskStatusEnded   = 4
)

const (
	HuTypeDianPao HuPaiType = iota
	HuTypeZiMo
	HuTypePei
)

const (
	SexTypeUnknown = 0
	SexTypeMale    = 1
	SexTypeFemale  = 2
)

const (
	UserTypeGuest    = 0
	UserTypeLaoBaShi = 1
)

const (
	FanPaiStepK91 = 0
	FanPaiStepK61 = 1
	FanPaiStepK41 = 2
	FanPaiStepK31 = 3
	FanPaiStepK21 = 4
)

const (
	FanPaiStatusKNotOpen1      = 0
	FanPaiStatusKOpenFailed1   = 1
	FanPaiStatusKOpenSuccessY1 = 2
	FanPaiStatusKOpenSuccessN1 = 3
	FanPaiStatusKNotOpen2      = 4
	FanPaiStatusKOpenFailed2   = 5
	FanPaiStatusKOpenSuccessY2 = 6
	FanPaiStatusKOpenSuccessN2 = 7
)

// OpType
const (
	OptypeIllegal = 0
	OptypeChu     = 1
	OptypePeng    = 2
	OptypeGang    = 3
	OptypeHu      = 4
	OptypePass    = 5

	OptyMoPai = 500 //摸牌
	//以下三种杠的分类主要用以解决上面的 OptypeGang分类不细致,导致抢杠等操作处理麻烦的问题
	//在判定时必须满足两条件 x % 10 == 4 && x >1000
	OptypeAnGang   = 1004
	OptypeMingGang = 1014
	OptypeBaGang   = 1024
)

// 番型
const (
	FanXingQingYiSe      = 1  // "清一色"
	FanXingQingQiDui     = 2  // "清七对"
	FanXingQingDaDui     = 3  // "清大对"
	FanXingQingDaiYao    = 4  // "清带幺"
	FanXingQingJiangDui  = 5  // "清将对"
	FanXingSuFan         = 6  // "素番"
	FanXingQiDui         = 7  // "七对"
	FanXingDaDui         = 8  // "大对子"
	FanXingQuanDaiYao    = 9  // "全带幺"
	FanXingJiangDui      = 10 // "将对"
	FanXingYaoJiuQiDui   = 11 // "幺九七对"
	FanXingQingLongQiDui = 12 // "清龙七对"
	FanXingLongQiDui     = 13 // "龙七对"
)

// 创建房间频道选项
const (
	ChannelOptionAll  = "allChannel"
	ChannelOptionHalf = "halfChannel"
)
