package sign_model

type GameRolesInfo struct {
	*ResponseInfo
	Data GameRolesList `json:"data,omitempty"`
}

type GameRolesList struct {
	List []GameRolesData `json:"list"`
}

type GameRolesData struct {
	UID        string `json:"game_uid"`
	Name       string `json:"nickname"`
	Level      int8   `json:"level"`
	Region     string `json:"region"`
	GameBiz    string `json:"game_biz"`
	IsChosen   bool   `json:"is_chosen"`
	IsOfficial bool   `json:"is_official"`
}
