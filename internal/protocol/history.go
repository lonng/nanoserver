package protocol

type HistoryListRequest struct {
	DeskID int64 `json:"desk_id"`
	Offset int   `json:"offset"`
	Count  int   `json:"count"`
}

type DeleteHistoryRequest struct {
	ID string `json:"id"` //历史ID
}
type HistoryByIDRequest struct {
	ID int64 `json:"id"` //历史ID
}

type HistoryLiteListRequest struct {
	DeskID int64 `json:"desk_id"`
	Offset int   `json:"offset"`
	Count  int   `json:"count"`
}

type HistoryLite struct {
	Id           int64  `json:"id"`
	DeskId       int64  `json:"desk_id"`
	Mode         int    `json:"mode"`
	BeginAt      int64  `json:"begin_at"`
	BeginAtStr   string `json:"begin_at_str"`
	EndAt        int64  `json:"end_at"`
	PlayerName0  string `json:"player_name0"`
	PlayerName1  string `json:"player_name1"`
	PlayerName2  string `json:"player_name2"`
	PlayerName3  string `json:"player_name3"`
	ScoreChange0 int    `json:"score_change0"`
	ScoreChange1 int    `json:"score_change1"`
	ScoreChange2 int    `json:"score_change2"`
	ScoreChange3 int    `json:"score_change3"`
}

type History struct {
	HistoryLite
	Snapshot string `json:"snapshot"`
}

type HistoryLiteListResponse struct {
	Code  int           `json:"code"`
	Total int64         `json:"total"` //总数量
	Data  []HistoryLite `json:"data"`
}

type HistoryListResponse struct {
	Code  int       `json:"code"`
	Total int64     `json:"total"` //总数量
	Data  []History `json:"data"`
}

type HistoryByIDResponse struct {
	Code int      `json:"code"`
	Data *History `json:"data"`
}
