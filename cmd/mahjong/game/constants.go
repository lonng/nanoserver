package game

type ScoreChangeType byte

const (
	ScoreChangeTypeAnGang ScoreChangeType = iota + 1
	ScoreChangeTypeBaGang
	ScoreChangeTypeHu
)

var scoreChangeTypeDesc = [...]string{
	ScoreChangeTypeAnGang: "下雨",
	ScoreChangeTypeBaGang: "刮风",
	ScoreChangeTypeHu:     "胡",
}

const (
	turnUnknown = 255 //最多可能只有4个方位
)

const (
	kCurPlayer = "player"
)

func (s ScoreChangeType) String() string {
	return scoreChangeTypeDesc[s]
}
