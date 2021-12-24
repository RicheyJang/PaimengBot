package genshin_draw

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/RicheyJang/PaimengBot/utils"
)

type DrawPool struct {
	Name         string `json:"name"` // [卡池名]
	Type         int    `json:"type"` // 卡池类型，参见上述
	EndTimestamp int64  `json:"end_timestamp"`

	Title  string `json:"title"`   // 卡池介绍中的标题
	PicURL string `json:"pic_url"` // 卡池图片URL

	Limit5 []string `json:"limit5,omitempty"` // UP 5星
	Limit4 []string `json:"limit4,omitempty"` // UP 4星

	Normal5Character []string `json:"normal5_character,omitempty"`
	Normal5Weapon    []string `json:"normal5_weapon,omitempty"`
	Normal4          []string `json:"normal4"`
	Normal3          []string `json:"normal3"`
}

// SavePools 保存池子信息入文件
func SavePools(tp int, pools []DrawPool) (err error) {
	prefix := getPrefixByType(tp)
	if len(prefix) == 0 {
		return fmt.Errorf("no such pool type")
	}
	for i, _ := range pools {
		pools[i].Name = prefix
		if len(pools) > 1 { // 多个同类池子，名称使用：角色1、角色2、...
			pools[i].Name += strconv.FormatInt(int64(i+1), 10)
		}
	}
	// 保存
	path := utils.PathJoin(GenshinDrawPoolDir, fmt.Sprintf("%v.json", prefix))
	res, err := json.MarshalIndent(pools, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, res, 0o644)
}

// LoadPoolsByPrefix 按前缀字符串获取该类型池子信息
func LoadPoolsByPrefix(prefix string) (pools []DrawPool) {
	if len(prefix) == 0 {
		return
	}
	path := utils.PathJoin(GenshinDrawPoolDir, fmt.Sprintf("%v.json", prefix))
	res, err := os.ReadFile(path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(res, &pools)
	return
}

// LoadPools 获取该类型池子信息
func LoadPools(tp int) (pools []DrawPool) {
	return LoadPoolsByPrefix(getPrefixByType(tp))
}

// UserInfo 用户信息（模拟抽卡）
type UserInfo struct {
	Last4 uint32
	Last5 uint32
}

func GetUserInfo(id int64) (u UserInfo) {
	key := fmt.Sprintf("genshin_draw.u%v", id)
	v, err := proxy.GetLevelDB().Get([]byte(key), nil)
	if err != nil {
		return
	}
	_ = json.Unmarshal(v, &u)
	return
}

func PutUserInfo(id int64, u UserInfo) error {
	key := fmt.Sprintf("genshin_draw.u%v", id)
	value, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return proxy.GetLevelDB().Put([]byte(key), value, nil)
}
