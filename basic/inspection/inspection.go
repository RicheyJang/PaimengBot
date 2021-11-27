package inspection

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/RicheyJang/PaimengBot/utils/consts"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "自检与清理",
	Usage: `
用法：
	自检：展示程序与环境状态
	清理临时数据：清空临时文件夹，并统计大小
`,
	IsSuperOnly: true,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"自检", "check", "状态"}).SetBlock(true).FirstPriority().Handle(selfCheck)
	proxy.OnCommands([]string{"清理临时数据"}).SetBlock(true).FirstPriority().Handle(cleanTemp)
}

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

func selfCheck(ctx *zero.Ctx) {
	env, self := "环境信息：\n", fmt.Sprintf("%v进程信息：\n", utils.GetBotNickname())
	// cpu相关信息
	cpuCount, err := cpu.Counts(false)
	if err != nil {
		log.Warn("cpu Counts err: ", err)
	}
	env += fmt.Sprintf("CPU：%v核心", cpuCount)
	cpuCount, err = cpu.Counts(true)
	if err != nil {
		log.Warn("cpu Counts err: ", err)
	}
	env += fmt.Sprintf(" %v逻辑处理器", cpuCount)

	// cpu使用率,每2秒一次，总共2次
	cpuPercent := float64(0)
	for i := 1; i <= 2; i++ {
		time.Sleep(time.Second)
		percent, err := cpu.Percent(time.Second, false)
		if err != nil || len(percent) == 0 {
			log.Warn("cpu percent err: ", err)
			percent = append(percent, 0.0)
		}
		cpuPercent += percent[0]
	}
	env += fmt.Sprintf(" 占用 %v\n", formatPercent(cpuPercent/2.0))

	// 物理内存信息
	memory, err := mem.VirtualMemory()
	if err != nil {
		log.Warn("virtual mem err: ", err)
		memory = &mem.VirtualMemoryStat{}
	}
	env += fmt.Sprintf("内存：占用 %v ( %v / %v )\n",
		formatPercent(memory.UsedPercent), formatBytesSize(memory.Free), formatBytesSize(memory.Total))

	// 机器启动时间戳
	bootTime, err := host.BootTime()
	if err != nil {
		log.Warn("boot time err: ", err)
	}
	env += fmt.Sprintf("启动时间：%v\n", formatTime(bootTime))

	//显示磁盘分区信息
	partitions, err := disk.Partitions(false)
	if err != nil {
		log.Warn("disk partitions err: ", err)
	}
	sumPart := len(partitions)
	var diskTotal, diskUsed uint64
	var diskPercent float64
	for _, part := range partitions {
		usage, err := disk.Usage(part.Mountpoint)
		if err != nil {
			log.Warn("disk usage err: ", err)
			sumPart -= 1
			continue
		}
		diskTotal += usage.Total
		diskUsed += usage.Used
		diskPercent += usage.UsedPercent
	}
	env += fmt.Sprintf("存储：总占用 %v ( %v / %v )\n",
		formatPercent(diskPercent/float64(sumPart)), formatBytesSize(diskUsed), formatBytesSize(diskTotal))

	//显示显络信息和IO
	//IOCounters, err := net2.IOCounters(true)
	//if err != nil {
	//	log.Warn("net io err: ", err)
	//}
	//for _, counter := range IOCounters {
	//	env += fmt.Sprintf("网卡 %v send:%v recv:%v\n", counter.Name, counter.BytesSent, counter.BytesRecv)
	//}

	//显示当前进程信息
	pid, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		log.Warn("process err: ", err)
	}
	pidName, _ := pid.Name()
	pidPercent, _ := pid.CPUPercent()
	pidMem, _ := pid.MemoryPercent()
	pidTime, _ := pid.CreateTime()
	pidConn, _ := pid.Connections()
	self += fmt.Sprintf("进程名：%v\nCPU占用：%v\n内存占用：%v\nGoroutine: %v\n启动时间：%v\n", pidName,
		formatPercent(pidPercent), formatPercent(float64(pidMem)), runtime.NumGoroutine(), formatTime(uint64(pidTime/1000)))
	self += fmt.Sprintf("网络连接：共%v条连接\n", len(pidConn))
	maxShowNum := 4
	if !utils.IsSuperUser(ctx.Event.UserID) || ctx.Event.SubType != "friend" { // 非超级用户好友私聊，保护网络连接信息，不展示
		maxShowNum = 0
	}
	for i, conn := range pidConn {
		if i >= maxShowNum {
			self += fmt.Sprintf("\t......\n")
			break
		}
		self += fmt.Sprintf("\t%v:%v->%v:%v\n",
			conn.Laddr.IP, conn.Laddr.Port, conn.Raddr.IP, conn.Raddr.Port)
	}

	// 发送消息
	msg := formResponse(env, self)
	ctx.SendChain(msg)
}

func formatTime(sec uint64) string {
	return time.Unix(int64(sec), 0).Format("2006-01-02 15:04:05")
}

func formatPercent(percent float64) string {
	return strconv.FormatFloat(percent, 'f', 2, 64) + "%"
}

func formatBytesSize(size uint64) string {
	fixs := []string{"b", "K", "M", "G", "T", "P"}
	fSize := float64(size)
	for i, fix := range fixs {
		if fSize <= 1024 || i == len(fixs)-1 {
			return strconv.FormatFloat(fSize, 'f', 2, 64) + fix
		}
		fSize /= 1024.0
	}
	return strconv.FormatUint(size, 10)
}

func formResponse(env, self string) message.MessageSegment {
	defaultInfo := env + "--------------------\n" + self
	fontSize := 20.0
	w, h := images.MeasureStringDefault(defaultInfo, fontSize, 1.3)
	w, h = w+20, h+40
	img := images.NewImageCtxWithBGRGBA255(int(w), int(h), 255, 255, 255, 255)
	// 贴环境状态描述文字
	err := img.PasteStringDefault(env, fontSize, 1.3, 10, 15, w-10)
	if err != nil {
		log.Warnf("PasteStringDefault err: %v", err)
		return message.Text(defaultInfo)
	}
	// 画线
	_, envHeight := images.MeasureStringDefault(env, fontSize, 1.3)
	img.PasteLine(5, 15+envHeight+4, w-5, 15+envHeight+4, 2, "gray")
	// 贴本进程状态描述文字
	err = img.PasteStringDefault(self, fontSize, 1.3, 10, 15+envHeight+20, w-10)
	if err != nil {
		log.Warnf("PasteStringDefault err: %v", err)
		return message.Text(defaultInfo)
	}
	// 生成图片文件
	msg, err := img.GenMessageAuto()
	if err != nil {
		log.Warnf("生成图片失败, err: %v", err)
		return message.Text(defaultInfo)
	}
	return msg
}
