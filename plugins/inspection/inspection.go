package inspection

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

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

// CheckEnvironment 生成主机环境信息
func CheckEnvironment() string {
	env := "环境信息：\n"
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
	return env
}

// CheckSelf 生成机器人自身信息
func CheckSelf(showNet bool) string {
	self := fmt.Sprintf("%v进程信息：\n", utils.GetBotNickname())

	// 当前进程信息
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

	// 生成网络信息内容
	maxShowNum := 4
	if !showNet { // 非超级用户好友私聊，保护网络连接信息，不展示
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
	return self
}

// CheckOnebot 机器人前端状态
func CheckOnebot(brief bool) string {
	// 登录状态
	ctx := utils.GetBotCtx()
	if ctx == nil {
		return "登陆号异常: 无法获取CTX"
	}

	rsp := ctx.CallAction("get_status", zero.Params{}).Data
	ok := rsp.Get("online").Bool() && rsp.Get("good").Bool()
	if brief {
		if ok {
			return "登录号状态正常"
		} else {
			return "登陆号异常！"
		}
	}

	bot := "登录号状态："
	// 版本信息
	version := ctx.GetVersionInfo()
	bot += "\n" + version.Get("app_name").String() + "状态"
	if ok {
		bot += "正常"
	} else {
		bot += "异常"
	}
	bot += "\n" + version.Get("app_name").String() + "版本：" + version.Get("app_version").String()
	if version.Get("app_name").String() != "go-cqhttp" {
		return bot
	}

	// 统计信息
	bot += fmt.Sprintf("\n收到的数据包总数: %s", rsp.Get("stat.PacketReceived").String())
	bot += fmt.Sprintf("\n发送的数据包总数: %s", rsp.Get("stat.PacketSent").String())
	bot += fmt.Sprintf("\n数据包丢失总数: %s", rsp.Get("stat.PacketLost").String())
	bot += fmt.Sprintf("\n接收消息总数: %s", rsp.Get("stat.MessageReceived").String())
	bot += fmt.Sprintf("\n发送消息总数: %s", rsp.Get("stat.MessageSent").String())
	bot += fmt.Sprintf("\n链接断开次数: %s", rsp.Get("stat.DisconnectTimes").String())
	bot += fmt.Sprintf("\n账号掉线次数: %s", rsp.Get("stat.LostTimes").String())
	return bot
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

func formResponse(texts ...string) message.MessageSegment {
	var defaultInfo string
	for i, str := range texts {
		if i != 0 {
			defaultInfo += "--------------------\n"
		}
		defaultInfo += str
	}
	// 初始化
	fontSize := 20.0
	w, h := images.MeasureStringDefault(defaultInfo, fontSize, 1.3)
	w, h = w+20, h+float64(len(texts))*22
	img := images.NewImageCtxWithBGRGBA255(int(w), int(h), 255, 255, 255, 255)
	height := 15.0
	// 贴文字
	for i, str := range texts {
		if i != 0 { // 画线
			img.PasteLine(5, height, w-5, height, 2, "gray")
			height += 10
		}
		err := img.PasteStringDefault(str, fontSize, 1.3, 10, height, w-10)
		if err != nil {
			log.Warnf("PasteStringDefault err: %v", err)
			return message.Text(defaultInfo)
		}
		_, subHeight := images.MeasureStringDefault(str, fontSize, 1.3)
		height += subHeight + 10
	}
	// 生成图片文件
	msg, err := img.GenMessageAuto()
	if err != nil {
		log.Warnf("生成图片失败, err: %v", err)
		return message.Text(defaultInfo)
	}
	return msg
}
