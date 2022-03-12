package mihoyo

import "github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_cookie"

func GetInitializaationPrompt() string {
	return `如何获取cookie或uid:
获取方法1、下载APP 应急食品
	cookie详细获取方法：打开应急食品 进入工具 进入管理米游社账号 添加账号 （登录你的账号） 长按你登录成功的账号即可复制
获取方法2、适用于有基础的同学
	打开 https://bbs.mihoyo.com/ys/ 登录后按F12，点击源代码禁用调试后，点击控制台输入document.cookie复制输出的内容即可
` + genshin_cookie.Info.Usage
}
