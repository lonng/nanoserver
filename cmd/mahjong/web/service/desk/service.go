package desk

import (
	"github.com/lonnng/nanoserver/db"
	"github.com/lonnng/nanoserver/internal/protocol"

	log "github.com/sirupsen/logrus"
	"time"
)

var supportOptions = `{
	"GET": "/v1/productions/",
	"DELETE": "/v1/productions/{id}",
	"OPTIONS": "/v1/productions/"
}`

var logger *log.Entry

type Service interface {
	DeskByID(id int64) (*protocol.Desk, error)

	DeskList(*protocol.DeskListRequest) ([]protocol.Desk, int64, error)
	DeleteDesk(id string) error
	GetOptions() string
}

type service struct {
	logger log.Logger
}

//NewService new a service for user
func NewService(l *log.Entry) Service {
	logger = l.WithField("service", "desk")
	return &service{}
}

func (s *service) DeskByID(id int64) (*protocol.Desk, error) {
	p, err := db.QueryDesk(id)
	if err != nil {
		return nil, err
	}
	return &protocol.Desk{
		Id:           p.Id,
		Creator:      p.Creator,
		Mode:         p.Mode,
		Round:        p.Round,
		DeskNo:       p.DeskNo,
		Player0:      p.Player0,
		Player1:      p.Player1,
		Player2:      p.Player2,
		Player3:      p.Player3,
		PlayerName0:  p.PlayerName0,
		PlayerName1:  p.PlayerName1,
		PlayerName2:  p.PlayerName2,
		PlayerName3:  p.PlayerName3,
		ScoreChange0: p.ScoreChange0,
		ScoreChange1: p.ScoreChange1,
		ScoreChange2: p.ScoreChange2,
		ScoreChange3: p.ScoreChange3,
		CreatedAt:    p.CreatedAt,
		DismissAt:    p.Creator,
	}, nil

}

func (s *service) DeskList(req *protocol.DeskListRequest) ([]protocol.Desk, int64, error) {
	//默认全部
	ps, total, err := db.DeskList(req.Player, req.Offset, req.Count)
	if err != nil {
		return nil, 0, err
	}
	list := make([]protocol.Desk, total)

	const (
		format = "2006-01-02 15:04:05"
	)

	for i, p := range ps {

		createdAtStr := time.Unix(p.CreatedAt, 0).Format(format)

		list[i] = protocol.Desk{
			Id:           p.Id,
			Creator:      p.Creator,
			Round:        p.Round,
			Mode:         p.Mode,
			DeskNo:       p.DeskNo,
			Player0:      p.Player0,
			Player1:      p.Player1,
			Player2:      p.Player2,
			Player3:      p.Player3,
			PlayerName0:  p.PlayerName0,
			PlayerName1:  p.PlayerName1,
			PlayerName2:  p.PlayerName2,
			PlayerName3:  p.PlayerName3,
			ScoreChange0: p.ScoreChange0,
			ScoreChange1: p.ScoreChange1,
			ScoreChange2: p.ScoreChange2,
			ScoreChange3: p.ScoreChange3,
			CreatedAt:    p.CreatedAt,
			CreatedAtStr: createdAtStr,
			DismissAt:    p.Creator,
		}
	}
	return list, int64(len(list)), nil
}

func (s *service) DeleteDesk(id string) error {
	return nil
}

func (*service) GetOptions() string {
	return supportOptions
}
