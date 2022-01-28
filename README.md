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

### 一般功能
- [x] 聊天 (chat)
- [x] 联系管理员 (contact)
- [x] 复读 (echo)
- [x] 自检 (inspection)
- [x] 功能使用统计(可分人分日) (statistic)
- [x] 回复撤回消息 (withdraw)
- [x] 点歌 (music)
- [x] 随机\随机数 (random)

### 原神相关
- [x] 今日可肝素材查询 (genshin_resource)
- [x] 模拟原神抽卡 (genshin_draw)

### 群功能
- [x] 设置入群欢迎 (welcome)

### 小游戏
- [x] 看图猜成语 (idioms)
- [ ] 多人五子棋

### 实用工具
- [x] 任意语种翻译(甚至文言文) (translate)
- [x] 纯小写缩写翻译 (hhsh)
- [x] 搜梗 (geng)
- [x] 识图搜番 (whatanime)
- [x] 疫情查询 (COVID)
- [x] 短链接还原 (short_url)
- [ ] ~~短链接生成~~(防止滥用，暂不提供)

### 好康的
- [x] 涩图 (pixiv)
- [x] Pixiv排行榜 (pixiv_rank)
- [ ] 查看Pixiv插图所有分P
- [ ] coser

</details>

## 安装与使用

参照详细文档：[派蒙Bot文档](https://richeyjang.github.io/PaimengBot)

如果安装或使用中遇到问题，或者有任何问题或建议想要讨论，欢迎加群[应急食品测试群(724694686)](https://qm.qq.com/cgi-bin/qm/qr?k=2u70XSTgORNbVzAnnsSYD2GLrelRuQC6&jump_from=webapi)

## 开发

#### 编译

本项目的编译基于go1.17+

此外，还需要64位的gcc，以保证cgo依赖包正常编译，若未安装64位gcc，可以从这里下载安装：http://tdm-gcc.tdragon.net/

#### 开发规范

参照开发文档：[派蒙Bot文档](https://richeyjang.github.io/PaimengBot)