package history

import (
	"time"

	//"triple/modules/errutil"
	"github.com/lonnng/nanoserver/db"
	"github.com/lonnng/nanoserver/internal/protocol"

	log "github.com/sirupsen/logrus"
)

var logger *log.Entry

const (
	format = "01-02 15:04:05"
)

var supportOptions = `{
	"GET": "/v1/productions/",
	"DELETE": "/v1/productions/{id}",
	"OPTIONS": "/v1/productions/"
}`

type Service interface {
	HistoryByID(id int64) (*protocol.History, error)

	HistoryList(*protocol.HistoryListRequest) ([]protocol.History, int64, error)
	HistoryLiteList(*protocol.HistoryLiteListRequest) ([]protocol.HistoryLite, int64, error)
	DeleteHistory(id string) error
	GetOptions() string
}

type service struct{}

//NewService new a service for user
func NewService(l *log.Entry) Service {
	logger = l.WithField("service", "history")
	return &service{}
}

func (s *service) HistoryByID(id int64) (*protocol.History, error) {
	p, err := db.QueryHistory(id)
	if err != nil {
		return nil, err
	}
	return &protocol.History{
		HistoryLite: protocol.HistoryLite{
			Id:           p.Id,
			DeskId:       p.DeskId,
			BeginAt:      p.BeginAt,
			Mode:         p.Mode,
			BeginAtStr:   time.Unix(p.BeginAt, 0).Format(format),
			EndAt:        p.EndAt,
			PlayerName0:  p.PlayerName0,
			PlayerName1:  p.PlayerName1,
			PlayerName2:  p.PlayerName2,
			PlayerName3:  p.PlayerName3,
			ScoreChange0: p.ScoreChange0,
			ScoreChange1: p.ScoreChange1,
			ScoreChange2: p.ScoreChange2,
			ScoreChange3: p.ScoreChange3,
		},
		Snapshot: p.Snapshot,
	}, nil

}

func (s *service) HistoryLiteList(req *protocol.HistoryLiteListRequest) ([]protocol.HistoryLite, int64, error) {
	//默认全部
	ps, total, err := db.QueryHistoriesByDeskID(req.DeskID)
	if err != nil {
		return nil, 0, err
	}
	list := make([]protocol.HistoryLite, total)
	for i, p := range ps {
		beginAtStr := time.Unix(p.BeginAt, 0).Format(format)
		list[i] = protocol.HistoryLite{
			Id:           p.Id,
			DeskId:       p.DeskId,
			Mode:         p.Mode,
			BeginAt:      p.BeginAt,
			BeginAtStr:   beginAtStr,
			EndAt:        p.EndAt,
			PlayerName0:  p.PlayerName0,
			PlayerName1:  p.PlayerName1,
			PlayerName2:  p.PlayerName2,
			PlayerName3:  p.PlayerName3,
			ScoreChange0: p.ScoreChange0,
			ScoreChange1: p.ScoreChange1,
			ScoreChange2: p.ScoreChange2,
			ScoreChange3: p.ScoreChange3,
		}
	}
	return list, int64(len(list)), nil
}

func (s *service) HistoryList(req *protocol.HistoryListRequest) ([]protocol.History, int64, error) {
	//默认全部
	ps, total, err := db.QueryHistoriesByDeskID(req.DeskID)
	if err != nil {
		return nil, 0, err
	}

	list := make([]protocol.History, total)
	for i, p := range ps {
		beginAtStr := time.Unix(p.BeginAt, 0).Format(format)
		list[i] = protocol.History{
			HistoryLite: protocol.HistoryLite{
				Id:           p.Id,
				Mode:         p.Mode,
				DeskId:       p.DeskId,
				BeginAt:      p.BeginAt,
				BeginAtStr:   beginAtStr,
				EndAt:        p.EndAt,
				PlayerName0:  p.PlayerName0,
				PlayerName1:  p.PlayerName1,
				PlayerName2:  p.PlayerName2,
				PlayerName3:  p.PlayerName3,
				ScoreChange0: p.ScoreChange0,
				ScoreChange1: p.ScoreChange1,
				ScoreChange2: p.ScoreChange2,
				ScoreChange3: p.ScoreChange3,
			},
			Snapshot: p.Snapshot,
		}
	}
	return list, int64(len(list)), nil
}

func (s *service) DeleteHistory(id string) error {
	return nil
}

func (*service) GetOptions() string {
	return supportOptions
}
