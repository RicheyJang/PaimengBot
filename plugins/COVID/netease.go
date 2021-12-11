package COVID

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/RicheyJang/PaimengBot/utils/client"
)

const covid163API = "https://c.m.163.com/ug/api/wuhan/app/data/list-total"

// 通过网易163接口获取疫情状况
func getCOVID19ConditionBy163(area string) ([]string, error) {
	c := client.NewHttpClient(nil)
	rsp, err := c.GetGJson(covid163API)
	if err != nil {
		return nil, err
	}
	if rsp.Get("code").Int() != 10000 {
		return nil, fmt.Errorf("code=%v, msg=%v", rsp.Get("code"), rsp.Get("msg"))
	}
	// 搜寻area
	if len(area) == 0 {
		area = "中国" // 默认区域为中国
	}
	minDepth, maxDepth := 1, uint(100)
	if strings.HasSuffix(area, "省") {
		minDepth = 2
		maxDepth = 2
		area = strings.TrimSuffix(area, "省")
	}
	if strings.HasSuffix(area, "市") {
		minDepth = 3
		area = strings.TrimSuffix(area, "市")
	}
	res := find163AreaTree(area, rsp.Get("data.areaTree").Array(), minDepth, maxDepth)
	return format163AreaCondition(res), nil
}

// 从areaTree及其children中寻找指定地区的疫情数据，并可以指定最大最小搜索深度
func find163AreaTree(name string, tree []gjson.Result, minDepth int, maxDepth uint) gjson.Result {
	if len(tree) == 0 || maxDepth == 0 {
		return gjson.Result{}
	}
	for _, area := range tree {
		if minDepth <= 1 && area.Get("name").String() == name { // 找到了
			return area
		}
		children := area.Get("children").Array()
		if len(children) > 0 {
			tmp := find163AreaTree(name, children, minDepth-1, maxDepth-1)
			if tmp.Exists() { // 在children中找到
				return tmp
			}
		}
	}
	return gjson.Result{}
}

// 格式化一个地区的疫情数据
func format163AreaCondition(area gjson.Result) (res []string) {
	if !area.Exists() {
		return nil
	}
	// 更新时间
	last := area.Get("lastUpdateTime")
	if last.Type != gjson.Null {
		res = append(res, fmt.Sprintf("更新时间：%v", last))
	}
	// 其它数据
	res = format163SingleNum(res, "确诊", area.Get("total.confirm"), area.Get("today.confirm"))
	beforeInputLen := len(res)
	res = format163SingleNum(res, "境外输入", area.Get("total.input"), area.Get("today.input"))
	if len(res) == beforeInputLen { // 没有境外输入数据，尝试从children中拿
		input := find163AreaTree("境外输入", area.Get("children").Array(), 1, 1)
		if input.Exists() { // children中有
			res = format163SingleNum(res, "境外输入", input.Get("total.confirm"), input.Get("today.confirm"))
		}
	}
	res = format163SingleNum(res, "无症状", area.Get("extData.noSymptom"), area.Get("extData.incrNoSymptom"))
	res = format163SingleNum(res, "疑似", area.Get("total.suspect"), area.Get("today.suspect"))
	res = format163SingleNum(res, "治愈", area.Get("total.heal"), area.Get("today.heal"))
	res = format163SingleNum(res, "死亡", area.Get("total.dead"), area.Get("today.dead"))
	return
}

// 格式化单项数据
func format163SingleNum(pre []string, title string, total gjson.Result, today gjson.Result) []string {
	if !total.Exists() || total.Int() == 0 {
		if today.Int() == 0 {
			return pre
		} else {
			return append(pre, fmt.Sprintf("%s：今日%+d", title, today.Int()))
		}
	}
	// 正常
	str := fmt.Sprintf("%s：%d", title, total.Int())
	if today.Exists() {
		str += fmt.Sprintf(" (今日%+d)", today.Int())
	}
	return append(pre, str)
}
