package protocol

type RegisterAppRequest struct {
	Name            string                       `json:"name"`             //应用名
	RedirectURI     string                       `json:"redirect_uri"`     //回调地址
	Extra           string                       `json:"extra"`            //应用描述
	CpID            string                       `json:"cp_id"`            //内容供应商ID
	ThirdProperties map[string]map[string]string `json:"third_properties"` //第三方属性
}

type RegisterAppResponse struct {
	Code int32   `json:"code"` //状态码
	Data AppInfo `json:"data"` //数据
}

type AppListRequest struct {
	CpID   string `json:"cp_id"`
	Offset int    `json:"offset"`
	Count  int    `json:"count"`
}
type AppListResponse struct {
	Code  int       `json:"code"`  //状态码
	Data  []AppInfo `json:"data"`  //应用列表
	Total int64     `json:"total"` //总数量
}

type AppInfoRequest struct {
	AppID string `json:"appid"` //应用ID
}

type AppInfoResponse struct {
	Code int32   `json:"code"` //状态码
	Data AppInfo `json:"data"` //数据
}

type DeleteAppRequest struct {
	AppID string `json:"appid"` //应用ID
}

type UpdateAppRequest struct {
	Type            int                          `json:"type"`             //更新类型, 1为重新生成appkey/appsecret, 2为更新其他内容
	AppID           string                       `json:"appid"`            //应用ID
	Name            string                       `json:"name"`             //应用名
	RedirectURI     string                       `json:"redirect_uri"`     //回调地址
	Extra           string                       `json:"extra"`            //应用描述
	ThirdProperties map[string]map[string]string `json:"third_properties"` //第三方属性
}
