package sign_model

type SignStateInfo struct {
	*ResponseInfo
	Data SignStateData `json:"data,omitempty"`
}

type SignStateData struct {
	Today        string `json:"today"`
	TotalSignDay int8   `json:"total_sign_day"`
	IsSign       bool   `json:"is_sign"`
	IsSub        bool   `json:"is_sub"`
	MonthFirst   bool   `json:"month_first"`
	FirstBind    bool   `json:"first_bind"`
}
