package protocol

type (
	ClubItem struct {
		Id        int64  `json:"id"`
		Name      string `json:"name"`
		Desc      string `json:"desc"`
		Member    int    `json:"member"`
		MaxMember int    `json:"maxMember"`
	}

	ClubListResponse struct {
		Code int        `json:"code"`
		Data []ClubItem `json:"data"`
	}

	ApplyClubRequest struct {
		ClubId int64 `json:"clubId"`
	}
)
