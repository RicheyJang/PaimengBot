package inspection

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "控制命令",
	Usage: `用于Bot的基本控制，仅限私聊
用法：
	自检：展示程序与环境状态
	清理临时数据：清空临时文件夹，并统计大小
	检查更新：检查Bot是否有更新
	关机：紧急关闭Bot程序

config-plugin配置项：
	inspection.timeout: 自动更新的超时时间，默认为5分钟
可以通过heartbeat系列配置项添加心跳检测并将心跳消息定期发送给监听人
若长时间没有收到心跳检测消息(且在监听时间段内)，说明机器人出现问题；但，非专业勿动：
	inspection.heartbeat.receiver: 心跳检测监听人ID列表
	inspection.heartbeat.interval: 心跳检测时间间隔
	inspection.heartbeat.period: 24小时制监听时间段，仅在该时间段内发送心跳消息`,
	IsSuperOnly: true,
}

const unknownVersion = "unknown"

var Version = unknownVersion

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnFullMatch([]string{"自检", "check", "状态"}, zero.OnlyPrivate).SetBlock(true).SecondPriority().Handle(selfCheckHandler)
	proxy.OnFullMatch([]string{"清理临时数据"}, zero.OnlyPrivate).SetBlock(true).SecondPriority().Handle(cleanTemp)
	proxy.OnFullMatch([]string{"检查更新"}, zero.OnlyPrivate).SetBlock(true).SecondPriority().Handle(updateHandler)
	proxy.OnFullMatch([]string{"关机"}, zero.OnlyPrivate).SetBlock(true).SecondPriority().Handle(shutdownHandler)
	proxy.OnFullMatch([]string{"重启"}, zero.OnlyPrivate).SetBlock(true).SecondPriority().Handle(restartHandler)
	proxy.AddConfig("timeout", "5m") // 默认超时 5分钟
	proxy.AddConfig("heartbeat.receiver", []int64{})
	proxy.AddConfig("heartbeat.interval", "1h")
	proxy.AddConfig("heartbeat.period", "9-22")
	manager.WhenConfigFileChange(heartbeatConfigHook)
}

// 清理临时文件
func cleanTemp(ctx *zero.Ctx) {
	usage := utils.PathSize(consts.TempRootDir)
	err := utils.RemovePath(consts.TempRootDir)
	if err != nil {
		log.Errorf("cleanTemp err: %v", err)
		ctx.Send("清理失败了...")
	} else {
		ctx.Send(fmt.Sprintf("成功清理%v大小的空间", formatBytesSize(usage)))
	}
}

// 自检
func selfCheckHandler(ctx *zero.Ctx) {
	if proxy.LockUser(0) {
		ctx.Send("自检中...")
		return
	}
	defer proxy.UnlockUser(0)

	msg := formResponse(CheckEnvironment(),
		CheckSelf(utils.IsSuperUser(ctx.Event.UserID) && ctx.Event.SubType == "friend"),
		CheckOnebot(false))
	ctx.SendChain(msg)
}

// 关机
func shutdownHandler(ctx *zero.Ctx) {
	if proxy.LockUser(0) {
		ctx.Send("请等待上一命令完成")
		return
	}
	defer proxy.UnlockUser(0)

	if !utils.GetConfirm(fmt.Sprintf("确定关闭%v？", utils.GetBotNickname()), ctx) {
		ctx.Send("已取消")
		return
	}
	ctx.Send("下次见咯")
	// 关机处理
	proxy.GetLevelDB().Close()
	log.Fatalf("被超级用户%v关闭", ctx.Event.UserID)
}

// 重启
func restartHandler(ctx *zero.Ctx) {
	if proxy.LockUser(0) {
		ctx.Send("请等待上一命令完成")
		return
	}
	defer proxy.UnlockUser(0)

	if !utils.GetConfirm(fmt.Sprintf("确定自动重启%v？这不一定会成功且无提示，建议手动重启", utils.GetBotNickname()), ctx) {
		ctx.Send("已取消")
		return
	}

	proc := exec.Command(os.Args[0], os.Args[1:]...)
	//proc.Stdin = os.Stdin
	proc.Stderr = os.Stderr
	proc.Stdout = os.Stdout
	err := proc.Start()
	if err != nil {
		log.Error("重启失败：", err)
		return
	}

	log.Info("NEW PID: ", proc.Process.Pid)
	pidErr := ioutil.WriteFile("./bot.pid", []byte(fmt.Sprintf("%d", proc.Process.Pid)), 0o644)
	if pidErr != nil {
		log.Errorf("save pid file error: %v", pidErr)
	}

	log.Fatal("旧进程退出")
}

// 升级
func updateHandler(ctx *zero.Ctx) {
	if proxy.LockUser(0) {
		ctx.Send("请等待上一命令完成")
		return
	}
	defer proxy.UnlockUser(0)

	if Version == unknownVersion {
		ctx.Send("当前为非官方发布版本，请自行升级：\nhttps://github.com/RicheyJang/PaimengBot")
		return
	}
	// 检查更新
	now, err := getLatestVersion()
	if err != nil || len(now) == 0 {
		log.Errorf("getLatestVersion err: %v", err)
		ctx.Send("检查失败了...")
		return
	}
	if now == Version {
		ctx.Send("现在已经是最新版啦")
		return
	}
	if !utils.GetConfirm(fmt.Sprintf("当前版本为%v，最新版本为%v，是否进行自动升级？", Version, now), ctx) {
		ctx.Send("已取消")
		return
	}
	// 执行更新
	if err = downloadAndReplace(now); err != nil {
		log.Errorf("downloadAndReplace err: %v", err)
		ctx.Send("失败了...")
		return
	}
	ctx.Send(fmt.Sprintf("更新成功，请自行重启%v以完成更新", utils.GetBotNickname()))
}
