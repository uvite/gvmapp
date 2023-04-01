package bot
// /策略现场
type GvmOptions struct {
	Exchange string `json:"exchange" form:"exchange"`
	Symbol   string `json:"symbol" form:"symbol"`
	Interval string `json:"interval" form:"interval"`
	Sname    string `json:"sname" form:"sname"`
	Code     string `json:"code" form:"code"`
}
