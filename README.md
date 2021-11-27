# 派蒙机器人

一个使用Onebot协议、基于ZeroBot的QQ娱乐机器人，支持30余种功能，可以大大丰富你的QQ体验或群活跃度，欸嘿。

推荐使用go-cqhttp作为QQ前端，推荐安装在Ubuntu中，推荐使用PostgreSQL作为数据库，但若你没有数据库基础也无需担心，派蒙Bot支持SQLite：一种轻型、无需配置的单文件式数据库。

安装与配置时需要少量的命令行操作基础，一般来说，计算机纯小白也可以在30分钟内完成配置，进行愉快的玩耍。

## 声明

本项目与米哈游(Mihoyo)公司旗下的原神(Genshin Impact)没有任何联系，仅为我个人开发出来作学习、娱乐所用，本项目**没有任何内容用于商业用途，没有任何收费项，注意辨别**，特此声明。

若有任何侵犯米哈游(Mihoyo)公司或原神(Genshin Impact)游戏权益的内容，请务必与我联系，我将立马删除整改，谢谢。

## 安装与使用

### 下载

从[Release](https://github.com/RicheyJang/PaimengBot/releases)页面下载最新版本的符合您操作系统的PaimengBot压缩包，在任意目录解压

**Linux 解压：**

```bash
tar -zxvf 压缩包名
```

### 首次运行

使用命令行进入上述文件夹，运行

**Windows：**

```bash
./paimeng.exe
```

**Linux：**

```bash
./paimeng
```

运行后，会自动生成配置文件，随后程序自动结束。

*此处应有图片*

### 配置GO-CQHTTP

go-cqhttp是一个用Go开发的QQ机器人前端，用于登录QQ机器人账号、收发消息等等，是派蒙Bot的必要依赖哦。

首先，从[go-cqhttp发布页](https://github.com/Mrs4s/go-cqhttp/releases)下载最新版本的符合您操作系统的go-cqhttp，放置在任意目录。

随后，使用命令行进入该目录，运行go-cqhttp。在首次运行时，它会要求你输入选择通信方式，选择 `2` ( Websocket 通信)，会生成一个配置文件**config.yml**。

*此处应有图片*

接下来，打开配置文件config.yml，将文件首部的account.uin配置项（冒号后的内容）修改为你所要使用的机器人QQ账号；将文件末尾处的ws配置中的host项修改为127.0.0.1，将port项修改为6700。

*此处应有图片*

最后，重新运行go-cqhttp。

若在这一步骤中出现问题，请参阅：[go-cqhttp文档](https://docs.go-cqhttp.org/guide/quick_start.html#%E5%9F%BA%E7%A1%80%E6%95%99%E7%A8%8B)。

### 配置数据库

目前，本项目支持SQLite、MySql、PostgreSQL这3种数据库，依据以下教程，**任选其一**进行配置即可。推荐使用PostgreSQL，会有更高的运行效率、更小的磁盘占用。

**但若你对任何数据库完全没有了解，或在配置过程中遇到了超出能力范围的问题，本项目可以选用SQLite：一种轻型、无需配置的单文件式数据库**

#### 1. SQLite

**Windows安装：**

一般来说，无需安装

**Linux安装**：

运行以下两条命令即可：

```bash
sudo apt-get install sqlite
sudo apt-get install libsqlite3-dev
```

**配置文件：**

随后，在派蒙Bot文件夹下的config-main.yaml配置文件中，修改db.type配置项为`sqlite`，修改db.name配置项为`./data/sqlite.db`即可。

*此处应有图片*

#### 2. PostgreSQL

**Window安装：**

1. 在[Postgresql官网](https://www.enterprisedb.com/downloads/postgres-postgresql-downloads)下载对应系统的Postgresql安装程序
2. 选择安装路径，一路next，中途会让你设置一下postgres用户的密码
3. 在安装目录下找到pgAdmin，使用pgAdmin连接数据库，创建连接，新建数据库即可

**Linux安装：**

自行百度，可以参考[教程](https://www.cnblogs.com/wwh/p/11605240.html)。

**配置文件：**

随后，在派蒙Bot文件夹下的config-main.yaml配置文件中，修改db.type配置项为`postgresql`，其它配置项：name修改为数据库名，host修改为`localhost`，port修改为`5432`，user修改为`postgres`或你此前自行创建的数据库用户名，passwd修改为你此前自行修改的数据库用户密码。

*此处应有图片*

#### 3. MySql

作为一种受众面最广的数据库，在此我就不说明怎么安装了……

**配置文件：**

在派蒙Bot文件夹下的config-main.yaml配置文件中，修改db.type配置项为`mysql`，其它配置项：name修改为数据库名，host修改为`localhost`，port修改为`3306`，user修改为`root`或你此前自行创建的数据库用户名，passwd修改为你此前自行修改的数据库用户密码。

*此处应有图片*

### 启动派蒙Bot

最后一项，在派蒙Bot文件夹下的config-main.yaml配置文件中，将superuser配置项（是一个YAML字符串数组类型）中，添加上你自己的QQ号（作为超级用户管理派蒙Bot）。

*此处应有图片*

类似**首次运行**，再次运行派蒙机器人。

### Enjoy it

## 开发

#### 编译

本项目的编译基于go1.17+

此外，还需要64位的gcc，以保证cgo依赖包正常编译，若未安装64位gcc，可以从这里下载安装：http://tdm-gcc.tdragon.net/