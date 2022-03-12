package constant

const (
	ActID      = "e202009291139501"
	AppVersion = "2.3.0"
	BaseUrl    = "https://webstatic.mihoyo.com/bbs/event/signin-ys/index.html"
	ClientType = "5"
	Salt       = "h8w582wxwgqvahcdkpvdhbh2w9casgfl"

	AcceptEncoding = "gzip, deflate"
	UserAgent      = "Mozilla/5.0 (Linux; Android 5.1.1; f103 Build/LYZ28N; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/52.0.2743.100 Safari/537.36 miHoYoBBS/" + AppVersion
	RefererUrl     = BaseUrl + "?bbs_auth_required=true&act_id=" + ActID + "&utm_source=bbs&utm_medium=mys&utm_campaign=icon"

	OpenApi = "https://api-takumi.mihoyo.com/"

	//GetUserGameRolesByCookie 获取账号信息
	GetUserGameRolesByCookie = "binding/api/getUserGameRolesByCookie?"

	//GetBbsSignRewardInfo 获取签到信息
	GetBbsSignRewardInfo = "event/bbs_sign_reward/info?"

	//PostSignInfo 签到
	PostSignInfo = "event/bbs_sign_reward/sign"
)

//文件名字
const (
	EnvFileName    = ".env"
	RecordFileName = "record.json"
	LogFileName    = "log.txt"
	CookieFileName = "cookie.txt"
)
