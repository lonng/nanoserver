package game

import (
	"fmt"

	"github.com/lonng/nanoserver/internal/game/mahjong"
)

type scoreChangeInfo struct {
	uid   int64 //谁的score发生了变化(+/-)
	score int   //变化了多少

	tileID int             //引起此种变化的tile id
	typ    ScoreChangeType //变化的类型
}

func (s *scoreChangeInfo) String() string {
	return fmt.Sprintf("Uid=%d, Score=%d, TileId=%s, Type=%s",
		s.uid, s.score, mahjong.TileFromID(s.tileID).String(), s.typ.String())
}
