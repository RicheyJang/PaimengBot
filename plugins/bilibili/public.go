package bilibili

import (
	"fmt"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"

	"github.com/tidwall/gjson"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var apiMap = make(map[string]string)

// SetAPIDefault 设置默认API地址
func SetAPIDefault(api string, value string) {
	apiMap[api] = value
}

// GetAPI 获取API地址
func GetAPI(api string) (url string) {
	return apiMap[api]
}

// 一些预定义
const (
	SearchTypeBangumi  = "media_bangumi"
	SearchTypeUser     = "bili_user"
	DynamicTypeShare   = 1 // 转发
	DynamicTypePic     = 2 // 图片动态
	DynamicTypeText    = 4 // 文字动态
	DynamicTypeVideo   = 8 // 视频动态
	LiveStatusClose    = 0 // 直播间关闭
	LiveStatusOpen     = 1 // 直播中
	LiveStatusCarousel = 2 // 直播间轮播中
)

type UserInfo struct {
	MID      int64  `json:"mid"`
	Name     string `json:"name"`
	Sex      string `json:"sex"`
	FaceURL  string `json:"face"`
	Sign     string `json:"sign"`
	Level    int    `json:"level"`
	Birthday string `json:"birthday"`
	// 与哔哩哔哩回包不一致的字段：
	Silence    bool  `json:"silence"` // 是否被封禁
	Fans       int64 `json:"fans"`    // 粉丝数（仅在搜索结果中提供）
	LiveRoomID int64 `json:"live_room_id"`
}

type DynamicInfo struct {
	ID   string    `json:"dynamic_id_str"`
	Type int       `json:"type"`
	Card string    `json:"card"`
	View int64     `json:"view"`
	Like int64     `json:"like"`
	Time time.Time `json:"timestamp"`
	// 与哔哩哔哩回包不一致的附加字段：
	Uname string `json:"uname"`
	BVID  string `json:"bvid"`
}

type BangumiInfo struct {
	MediaID     int64  `json:"media_id"`
	SeasonID    int64  `json:"season_id"`
	Title       string `json:"title"`
	OrgTitle    string `json:"org_title"`
	Areas       string `json:"areas"`
	Description string `json:"desc"`
	Styles      string `json:"styles"`
	EPSize      int    `json:"ep_size"`
	CoverURL    string `json:"cover"`
	// 与哔哩哔哩回包不一致的字段：
	Score float64 `json:"score"`
}

type BangumiEPInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"index"`
	IndexShow string `json:"index_show"`
}

type BangumiLatestInfo struct {
	MediaID  int64         `json:"media_id"`
	SeasonID int64         `json:"season_id"`
	Title    string        `json:"title"`
	Areas    string        `json:"areas"`
	CoverURL string        `json:"cover"`
	NewEP    BangumiEPInfo `json:"new_ep"`
	URL      string        `json:"share_url"`
	// 与哔哩哔哩回包不一致的字段：
	Score float64 `json:"score"`
}

type LiveRoomInfo struct {
	ID       int64    `json:"room_id"`
	ShortID  int64    `json:"short_id"`
	Title    string   `json:"title"`
	Status   int      `json:"live_status"`
	CoverURL string   `json:"cover"`
	Anchor   UserInfo `json:"anchor_info"`
}

func (d DynamicInfo) IsVideo() bool {
	return d.Type == DynamicTypeVideo
}

func (d DynamicInfo) VideoTitle() string {
	return gjson.Get(d.Card, "title").String()
}

func (l LiveRoomInfo) IsOpen() bool {
	return l.Status == LiveStatusOpen
}

func (u UserInfo) GenMessage(index int) message.Message {
	str := u.Name + "\n"
	if index > 0 {
		str = "[" + fmt.Sprintf("%d", index) + "] " + str
	}
	str += "粉丝数：" + fmt.Sprintf("%d", u.Fans) + "\n"
	str += "等级：lv" + fmt.Sprintf("%d", u.Level)
	return message.Message{message.Text(str)}
}

func (b BangumiInfo) GenMessage(index int) message.Message {
	str := b.Title + "\n"
	if index > 0 {
		str = "[" + fmt.Sprintf("%d", index) + "] " + str
	}
	str += "番剧ID：" + fmt.Sprintf("%d", b.MediaID) + "\n"
	str += "地区：" + b.Areas + "\n"
	if utils.StringRealLength(b.Description) > 50 {
		str += "简介：" + string([]rune(b.Description)[:50]) + "..."
	} else {
		str += "简介：" + b.Description
	}
	return message.Message{message.Text(str)}
}

// Client bilibili客户端（请求器）
type Client struct {
	*client.HttpClient
}

func NewClient() *Client {
	c := client.NewHttpClient(nil)
	c.SetUserAgent()
	c.SetHeader("Referer", "https://www.bilibili.com/")
	return &Client{
		HttpClient: c,
	}
}

// Search Bilibili搜索相关
type Search struct {
	c *Client
}

func NewSearch() *Search {
	return &Search{c: NewClient()}
}

func (s *Search) Type(searchType string, keyword string, additionalKV ...string) (gjson.Result, error) {
	if s == nil || s.c == nil {
		return gjson.Result{}, fmt.Errorf("search or client is nil")
	}
	api := GetAPI("search.type")
	api = api + "?search_type=" + searchType + "&keyword=" + keyword
	for i := 0; i+1 < len(additionalKV); i += 2 {
		api += fmt.Sprintf("&%s=%s", additionalKV[i], additionalKV[i+1])
	}
	rsp, err := s.c.GetGJson(api)
	if err == nil && rsp.Get("code").Int() != 0 {
		return gjson.Result{}, fmt.Errorf("bilibili error: %s", rsp.Get("message").String())
	}
	return rsp, err
}

func (s *Search) User(keyword string) ([]UserInfo, error) {
	rsp, err := s.Type(SearchTypeUser, keyword, "order", "fans")
	if err != nil {
		return nil, err
	}
	var users []UserInfo
	for _, v := range rsp.Get("data.result").Array() {
		face := v.Get("upic").String()
		if len(face) > 0 && !strings.HasPrefix(face, "http") {
			face = "https:" + face
		}
		users = append(users, UserInfo{
			MID:     v.Get("mid").Int(),
			Name:    v.Get("uname").String(),
			FaceURL: face,
			Sign:    v.Get("usign").String(),
			Level:   int(v.Get("level").Int()),
			Fans:    v.Get("fans").Int(),

			Sex:        "",
			Birthday:   "",
			Silence:    false,
			LiveRoomID: 0,
		})
	}
	return users, nil
}

func (s *Search) Bangumi(keyword string) ([]BangumiInfo, error) {
	rsp, err := s.Type(SearchTypeBangumi, keyword)
	if err != nil {
		return nil, err
	}
	var bangumi []BangumiInfo
	for _, v := range rsp.Get("data.result").Array() {
		bangumi = append(bangumi, BangumiInfo{
			MediaID:     v.Get("media_id").Int(),
			SeasonID:    v.Get("season_id").Int(),
			Title:       replaceEM(v.Get("title").String()),
			OrgTitle:    replaceEM(v.Get("org_title").String()),
			Description: v.Get("desc").String(),
			Styles:      v.Get("styles").String(),
			Score:       v.Get("media_score.score").Float(),
			EPSize:      int(v.Get("ep_size").Int()),
			CoverURL:    v.Get("cover").String(),
			Areas:       v.Get("areas").String(),
		})
	}
	return bangumi, nil
}

func replaceEM(org string) string {
	str := strings.ReplaceAll(org, `<em class="keyword">`, "")
	str = strings.ReplaceAll(str, `<em class=\"keyword\">`, "")
	return strings.ReplaceAll(str, `</em>`, "")
}

// Bangumi Bilibili番剧相关
type Bangumi struct {
	c *Client
}

func NewBangumi() *Bangumi {
	return &Bangumi{c: NewClient()}
}

func (b *Bangumi) ByMDID(mediaID int64) (BangumiLatestInfo, error) {
	if b == nil || b.c == nil {
		return BangumiLatestInfo{}, fmt.Errorf("bangumi or client is nil")
	}
	api := GetAPI("bangumi.mdid")
	api = api + "?media_id=" + fmt.Sprintf("%d", mediaID)
	rsp, err := b.c.GetGJson(api)
	if err != nil {
		return BangumiLatestInfo{}, err
	}
	if rsp.Get("code").Int() != 0 {
		return BangumiLatestInfo{}, fmt.Errorf("bilibili error: %s", rsp.Get("message").String())
	}
	rsp = rsp.Get("result.media")
	return BangumiLatestInfo{
		MediaID:  rsp.Get("media_id").Int(),
		SeasonID: rsp.Get("season_id").Int(),
		Title:    rsp.Get("title").String(),
		Areas:    rsp.Get("areas.0.name").String(),
		CoverURL: rsp.Get("cover").String(),
		NewEP: BangumiEPInfo{
			ID:        rsp.Get("new_ep.id").Int(),
			Name:      rsp.Get("new_ep.index").String(),
			IndexShow: rsp.Get("new_ep.index_show").String(),
		},
		URL:   rsp.Get("share_url").String(),
		Score: rsp.Get("rating.score").Float(),
	}, nil
}

// User Bilibili用户(up主)相关
type User struct {
	c    *Client
	id   int64
	info *UserInfo
}

func NewUser(ID int64) *User {
	return &User{c: NewClient(), id: ID}
}

func (u *User) Info() (UserInfo, error) {
	if u == nil || u.c == nil {
		return UserInfo{}, fmt.Errorf("user or client is nil")
	}
	if u.info != nil {
		return *u.info, nil
	}
	api := GetAPI("user.info")
	api = api + "?mid=" + fmt.Sprintf("%d", u.id)
	rsp, err := u.c.GetGJson(api)
	if err != nil {
		return UserInfo{}, err
	}
	if rsp.Get("code").Int() != 0 {
		return UserInfo{}, fmt.Errorf("bilibili error: %s", rsp.Get("message").String())
	}
	rsp = rsp.Get("data")
	user := UserInfo{
		MID:        rsp.Get("mid").Int(),
		Name:       rsp.Get("name").String(),
		Sex:        rsp.Get("sex").String(),
		FaceURL:    rsp.Get("face").String(),
		Sign:       rsp.Get("sign").String(),
		Level:      int(rsp.Get("level").Int()),
		Birthday:   rsp.Get("birthday").String(),
		Silence:    rsp.Get("silence").Int() == 1,
		LiveRoomID: rsp.Get("live_room.roomid").Int(),
	}
	if user.MID > 0 {
		u.info = &user
	}
	return user, nil
}

func (u *User) Dynamics(offset int, hasTop bool) ([]DynamicInfo, string, error) {
	if u == nil || u.c == nil {
		return nil, "0", fmt.Errorf("user or client is nil")
	}
	api := GetAPI("user.dynamic")
	api = api + "?host_uid=" + fmt.Sprintf("%d", u.id) +
		"&offset_dynamic_id=" + fmt.Sprintf("%d", offset) +
		"&need_top=" + fmt.Sprintf("%v", hasTop)
	rsp, err := u.c.GetGJson(api)
	if err != nil {
		return nil, "0", err
	}
	if rsp.Get("code").Int() != 0 {
		return nil, "0", fmt.Errorf("bilibili error: %s", rsp.Get("message").String())
	}
	var dynamics []DynamicInfo
	for _, v := range rsp.Get("data.cards").Array() {
		dynamics = append(dynamics, DynamicInfo{
			ID:   v.Get("desc.dynamic_id_str").String(),
			Type: int(v.Get("desc.type").Int()),
			Card: strings.ReplaceAll(v.Get("card").String(), `\/`, "/"),
			View: v.Get("desc.view").Int(),
			Like: v.Get("desc.like").Int(),
			Time: parseTimestampAuto(v.Get("desc.timestamp").Int()),

			Uname: v.Get("desc.user_profile.info.uname").String(),
			BVID:  v.Get("desc.bvid").String(),
		})
	}
	return dynamics, rsp.Get("data.next_offset").String(), nil
}

func parseTimestampAuto(stamp int64) time.Time {
	if stamp >= 1e15 { // 微秒
		return time.UnixMicro(stamp)
	} else if stamp >= 1e12 { // 毫秒
		return time.UnixMilli(stamp)
	}
	return time.Unix(stamp, 0)
}

// LiveRoom 直播相关
type LiveRoom struct {
	c  *Client
	id int64
}

func NewLiveRoom(id int64) *LiveRoom {
	return &LiveRoom{
		c:  NewClient(),
		id: id,
	}
}

func (l *LiveRoom) Info() (LiveRoomInfo, error) {
	if l == nil || l.c == nil {
		return LiveRoomInfo{}, fmt.Errorf("liveroom or client is nil")
	}
	api := GetAPI("live.info")
	api = api + "?room_id=" + fmt.Sprintf("%d", l.id)
	rsp, err := l.c.GetGJson(api)
	if err != nil {
		return LiveRoomInfo{}, err
	}
	if rsp.Get("code").Int() != 0 {
		return LiveRoomInfo{}, fmt.Errorf("bilibili error: %s", rsp.Get("message").String())
	}
	rsp = rsp.Get("data")
	return LiveRoomInfo{
		ID:       rsp.Get("room_info.room_id").Int(),
		ShortID:  rsp.Get("room_info.short_id").Int(),
		Title:    rsp.Get("room_info.title").String(),
		Status:   int(rsp.Get("room_info.live_status").Int()),
		CoverURL: rsp.Get("room_info.cover").String(),
		Anchor: UserInfo{
			MID:     rsp.Get("room_info.uid").Int(),
			Name:    rsp.Get("anchor_info.base_info.uname").String(),
			Sex:     rsp.Get("anchor_info.base_info.gender").String(),
			FaceURL: rsp.Get("anchor_info.base_info.face").String(),
			// 以下信息无效
			Sign:     "",
			Level:    0,
			Birthday: "",
			Silence:  false,
		},
	}, nil
}
