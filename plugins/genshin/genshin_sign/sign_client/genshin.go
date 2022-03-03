package sign_client

import (
	"fmt"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_sign/sign_model"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_sign/sign_util"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_sign/sign_util/constant"
	uuid "github.com/satori/go.uuid"

	"net/http"
)

type GenshinClient struct {
	*HTTPClient
}

// modifyHeader 添加访问mhy原神BBS需要的基本消息头
func modifyHeader(cookie string) func(req *http.Request) error {
	return func(req *http.Request) error {
		AddBasicHeader()(req)
		AddCookieHeader(cookie)(req)
		req.Header.Add("x-rpc-device_id", uuid.NewV4().String())
		return nil
	}
}

// addExtraGenshinHeader 添加访问mhy原神BBS需要的额外消息头
func addExtraGenshinHeader() func(req *http.Request) error {
	return func(req *http.Request) error {
		req.Header.Add("x-rpc-client_type", constant.ClientType)
		req.Header.Add("x-rpc-app_version", constant.AppVersion)
		req.Header.Add("DS", sign_util.GetDs())
		return nil
	}
}

// GetMhyURL 组合URL
func GetMhyURL(path, parameters string) (url string) {
	if parameters == "" {
		url = fmt.Sprintf("%s%s", constant.OpenApi, path)
	} else {
		url = fmt.Sprintf("%s%s%s", constant.OpenApi, path, parameters)
	}
	return
}

func NewGenshinClient() (g *GenshinClient) {
	g = &GenshinClient{
		HTTPClient: NewClient(),
	}
	return
}

// GetUserGameRoles 获取用户游戏角色
func (g *GenshinClient) GetUserGameRoles(cookie string) (rolesList []sign_model.GameRolesData) {
	url := GetMhyURL(constant.GetUserGameRolesByCookie, "game_biz=hk4e_cn")

	var info sign_model.GameRolesInfo
	err := g.SendGetMessage(url, nil, &info, modifyHeader(cookie))

	if err != nil {
		//log.Error("unable to send http massage.", err)
		return nil
	}

	switch info.Code {
	case 0:
		//log.Debug("get user game roles success.")
		break
	default:
		//log.Error("get user game roles error(%v). request failure.", info.Code, info.Msg, info.Data)
		break
	}
	return info.Data.List
}

// GetSignStateInfo 获取签到信息
func (g *GenshinClient) GetSignStateInfo(cookie string, roles sign_model.GameRolesData) (data *sign_model.SignStateData) {
	url := GetMhyURL(
		constant.GetBbsSignRewardInfo,
		fmt.Sprintf("act_id=%s&region=%s&uid=%s", constant.ActID, roles.Region, roles.UID),
	)

	var info sign_model.SignStateInfo
	err := g.SendGetMessage(url, nil, &info, modifyHeader(cookie))

	if err != nil {
		//log.Error("unable to send http massage.", err)
		return nil
	}
	switch info.Code {
	case 0:
		//log.Debug("get sign reward info success.")
		break
	default:
		//log.Error("get sign reward info error(%v). request failure.", info.Code, info.Msg, info.Data)
		break
	}
	return &info.Data
}

// Sign 进行签到
func (g *GenshinClient) Sign(cookie string, roles sign_model.GameRolesData) bool {
	data := map[string]string{
		"act_id": constant.ActID,
		"region": roles.Region,
		"uid":    roles.UID,
	}
	url := GetMhyURL(constant.PostSignInfo, "")

	var info sign_model.SignResponseInfo
	err := g.SendPostMessage(url, data, &info, modifyHeader(cookie), addExtraGenshinHeader())
	if err != nil {
		//log.Error("unable to send http massage.", err)
		return false
	}

	switch info.Code {
	case 0:
		//log.Debug("roles(%v:%v) sign success.", roles.UID, roles.Name)
		return true
	case -5003:
		//log.Debug("roles(%v:%v) sign info(%v). request failure. %v", roles.UID, roles.Name, info.Code, info.Msg, info.Data)
		return true
	default:
		//log.Error("roles(%v:%v) sign error(%v). request failure. %v", roles.UID, roles.Name, info.Code, info.Msg, info.Data)
		return false
	}
}
