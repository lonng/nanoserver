package protocol

type RegisterAgentRequest struct {
	Name     string `json:"name"`
	Account  string `json:"account"`
	Password string `json:"password"`
	Extra    string `json:"extra"`
}

type AgentLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AgentDetail struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	Account   string `json:"account"`
	CardCount int64  `json:"card_count"`
	CreateAt  int64  `json:"create_at"`
}

type AgentLoginResponse struct {
	Code   int         `json:"code"`
	Token  string      `json:"token"`
	Detail AgentDetail `json:"detail"`
}

type AgentListResponse struct {
	Code   int           `json:"code"`
	Agents []AgentDetail `json:"agents"`
	Total  int64         `json:"total"`
}

type RechargeDetail struct {
	PlayerId  int64  `json:"player_id"`
	Extra     string `json:"extra"`
	CreateAt  int64  `json:"create_at"`
	CardCount int64  `json:"card_count"`
}

type RechargeListResponse struct {
	Code      int              `json:"code"`
	Recharges []RechargeDetail `json:"recharges"`
	Total     int64            `json:"total"`
}
