# 派蒙机器人

一个使用Onebot协议、基于ZeroBot的QQ娱乐机器人，支持20余种功能，可以大大丰富你的QQ体验或群活跃度，欸嘿。

推荐使用go-cqhttp作为QQ前端，推荐安装在Ubuntu中，推荐使用PostgreSQL作为数据库，但若你没有数据库基础也无需担心，派蒙Bot支持SQLite：一种轻型、无需配置的单文件式数据库。

安装与配置时需要少量的命令行操作基础，一般来说，计算机纯小白也可以在30分钟内完成配置，进行愉快的玩耍。

## 声明

本项目与米哈游(Mihoyo)公司旗下的原神(Genshin Impact)没有任何联系，仅为我个人开发出来作学习、娱乐所用，本项目**没有任何内容用于商业用途，没有任何收费项，注意辨别**，特此声明。

若有任何侵犯米哈游(Mihoyo)公司或原神(Genshin Impact)游戏权益的内容，请务必与我联系，我将立马删除整改，谢谢。

## 功能

<details>
<summary>功能列表</summary>

行尾括号内为插件Key，对应着配置文件config-plugin.yaml中各个插件的根配置项key

### 基本功能
- [x] 权限管理与鉴权 (auth)
- [x] 功能开关与封(解)禁 (ban)
- [x] 加群\好友申请事件处理\推送 (event)
- [x] 帮助 (help)
- [x] 功能CD限流 (limiter)
- [x] 用户昵称系统 (nickname)
- [x] 签到与财富 (sc)

### 一般功能
- [x] 聊天\自定义问答 (chat)
- [x] 定期提醒 (note)
- [x] 联系管理员 (contact)
- [x] 复读 (echo)
- [x] 自检 (inspection)
- [x] 功能使用统计(可分人分日) (statistic)
- [x] 网易云评论 (netease)
- [x] 点歌 (music)
- [x] 随机\随机数 (random)
- [x] 漂流瓶 (bottle)
- [x] 戳一戳 (poke)

### 原神相关
- [x] 今日可肝素材查询 (genshin_resource)
- [x] 模拟原神抽卡 (genshin_draw)
- [x] 米游社管理 (genshin_cookie)
- [x] 米游社签到 (genshin_sign)
- [x] 原神便笺查询 (genshin_query)

### 实用工具
- [x] B站订阅、自动推送 (bilibili)
- [x] 任意语种翻译(甚至文言文) (translate)
- [x] 纯小写缩写翻译 (hhsh)
- [x] 搜梗 (geng)
- [x] 识图搜番 (whatanime)
- [x] 疫情查询 (COVID)
- [x] 短链接还原 (short_url)
- [x] 天气 (weather)
- [x] GitHub查询 (github)
- [x] 混合表情 (emoji_mix)

### 群功能
- [x] 群管理:快捷禁言/踢人 (admin)
- [x] 撤回消息 (withdraw)
- [x] 设置入群欢迎 (welcome)
- [x] 关键词撤回 (keyword)

### 小游戏
- [x] 看图猜成语 (idioms)
- [ ] 谁是卧底
- [ ] 文字RPG

### 好康的
- [x] 涩图 (pixiv)
- [x] Pixiv排行榜 (pixiv_rank)
- [x] Pixiv搜索 (pixiv_query)
- [ ] coser
- [ ] 自定义图库

### 可选插件

若想启用这些插件，请自行下载源码取消掉cmd/main.go内的可选插件注释，[**自行编译**](#编译)

- [x] OSU查询 (HiOSU)

</details>

## 安装与使用

参照详细文档：[派蒙Bot文档](https://richeyjang.github.io/PaimengBot)

如果安装或使用中遇到问题，或者有任何问题或建议想要讨论，总之欢迎加群[应急食品测试群(724694686)](https://qm.qq.com/cgi-bin/qm/qr?k=2u70XSTgORNbVzAnnsSYD2GLrelRuQC6&jump_from=webapi)

开发不易，如果感觉还不错，就在右上角点个star好啦，谢谢

## FAQ

**为什么在私聊中使用正常，在群聊中没反应？**

为了防止派蒙Bot在群聊中乱答话，特将部分可能产生歧义的功能设计为在群聊中调用需要加上派蒙Bot的名字前缀。例如：`派蒙帮助`、`派蒙关闭复读`，详细参见[issues#1](https://github.com/RicheyJang/PaimengBot/issues/1)。

而在私聊中则无需加上名字前缀。

**详情帮助中的方括号是什么意思？**

在部分功能的详细帮助中，常看见一条命令后带有方括号包裹的描述性文字，它们代表一个个参数、占位符，需要将特定的内容置于该位置上。

例如：好友群组管理插件中的`退群 [群号]`命令，在`[群号]`位置上则应该放置一个群号，比如`退群 724694686`，派蒙Bot便知道需要退出群号为724694686的群。

此外，部分方括号后还带有`?`,`*`,`+`等符号，`?`表示该参数可填可不填，`*`代表可以有任意个该参数，以空格分隔即可，`+`代表至少需要有一个该参数，同样以空格分隔即可。

**为什么拉派蒙Bot入群时，它自动退群了？**

为了防止有用户恶意拉群，特将派蒙Bot设计为非超级用户拉群时自动退群，但由于go-cqhttp端并没有提供拉群人的QQ号，因此可能会出现超级用户拉群也会自动退群的情况。

此时，仅需在私聊中对派蒙Bot说：`同意群邀请 [群号]`，随后再拉一遍即可成功入群。

## 开发

#### 编译

本项目使用纯Go语言实现，编译基于[go1.17+](https://golang.google.cn/dl/)

```bash
go get ./...

go build ./cmd/main.go
```

#### 开发文档

派蒙Bot作为一个较为完毕的机器人后端框架，提供了**插件式**集中管理和许多固有能力，参见开发文档。

[派蒙Bot开发文档](https://richeyjang.github.io/PaimengBot/develop/)