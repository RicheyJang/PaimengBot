package PaimengBot

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	"github.com/fsnotify/fsnotify"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	pflag.StringP("server", "s", "ws://127.0.0.1:6700/", "the websocket server address")
	pflag.StringSliceP("superuser", "u", []string{}, "all superusers' id")
	pflag.StringP("nickname", "n", "派蒙", "the bot's nickname")
	pflag.StringP("log", "l", "info", "the level of logging")
	pflag.BoolP("daemon", "d", false, "run the bot as a service")
	pflag.Parse()
	// 从命令行读取
	_ = viper.BindPFlag("superuser", pflag.Lookup("superuser"))
	_ = viper.BindPFlag("nickname", pflag.Lookup("nickname"))
	// 后端配置
	_ = viper.BindPFlag("server.address", pflag.Lookup("server"))
	viper.SetDefault("server.token", "")
	// 日志配置
	_ = viper.BindPFlag("log.level", pflag.Lookup("log"))
	viper.SetDefault("log.date", 30)
	// 数据库配置
	viper.SetDefault("db.type", "postgresql")
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 5432)
	viper.SetDefault("db.user", "username")
	viper.SetDefault("db.passwd", "password")
	viper.SetDefault("db.name", "database")
	// 其它配置
	viper.SetDefault("tmp.maxcount", 1000)        // 同种类临时文件同时存在的最大数量
	viper.SetDefault(consts.AlwaysCallKey, false) // 是否可以自由调用（完全去除onlytome），不支持热更新
	// 此init会在manager.common前被调用，随后manager.common.init调用DoPreWorks
}

// DoPreWorks 进行全局初始化工作
func DoPreWorks() {
	// 读取主配置
	err := flushMainConfig(consts.DefaultConfigDir, consts.MainConfigFileName)
	if err != nil {
		log.Fatal("FlushMainConfig err: ", err)
		return
	}
	// 初始化日志
	err = setupLogger()
	if err != nil {
		log.Fatal("setupLogger err: ", err)
		return
	}
	// 检查是否以服务模式启动
	CheckDaemon()
}

// 设置日志
func setupLogger() error {
	// 日志等级
	log.SetLevel(log.InfoLevel)
	if l, ok := flagLToLevel[strings.ToLower(viper.GetString("log.level"))]; ok {
		log.SetLevel(l)
	}
	// 日志格式
	log.SetFormatter(&utils.SimpleFormatter{})
	// 日志滚动切割
	logf, err := rotatelogs.New(
		utils.PathJoin(consts.DefaultLogDir, "bot-%Y-%m-%d.log"),
		rotatelogs.WithLinkName(utils.PathJoin(consts.DefaultLogDir, "bot.log")),
		rotatelogs.WithMaxAge(time.Duration(viper.GetInt("log.date"))*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Error("Get rotate logs err: ", err)
		return err
	}
	// 日志输出
	var stdOuter io.Writer = os.Stdout
	logWriter := io.MultiWriter(stdOuter, logf)
	log.SetOutput(logWriter) // logrus 设置日志的输出方式
	return nil
}

var flagLToLevel = map[string]log.Level{
	"debug":   log.DebugLevel,
	"info":    log.InfoLevel,
	"warn":    log.WarnLevel,
	"warning": log.WarnLevel,
	"error":   log.ErrorLevel,
}

// 从文件和命令行中刷新所有主配置，若文件不存在将会把配置写入该文件
func flushMainConfig(configPath string, configFileName string) error {
	// 从文件读取
	viper.AddConfigPath(configPath)
	viper.SetConfigFile(configFileName)
	fullPath := utils.PathJoin(configPath, configFileName)
	//fileType := filepath.Ext(fullPath)
	//viper.SetConfigType(fileType)
	if utils.FileExists(fullPath) { // 配置文件已存在：合并自配置文件后重新写入
		err := viper.MergeInConfig()
		if err != nil {
			log.Error("FlushMainConfig error in MergeInConfig err: ", err)
			return err
		}
		_ = viper.WriteConfigAs(fullPath)
	} else { // 配置文件不存在：写入配置
		err := viper.SafeWriteConfigAs(fullPath)
		if err != nil {
			log.Error("FlushMainConfig error in SafeWriteConfig err: ", err)
			return err
		}
		log.SetFormatter(&utils.SimpleFormatter{})
		log.Fatalf("初始化配置文件%v完成，请对该配置文件进行配置后，重启本程序", configFileName)
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) { // 配置文件发生变更之后会调用的回调函数
		zero.BotConfig.SuperUsers = viper.GetStringSlice("superuser")
		zero.BotConfig.NickName = []string{viper.GetString("nickname")}
		_ = setupLogger()
		log.Infof("reload main config from %v", e.Name)
	})
	return nil
}

// CheckDaemon 检查是否需要以服务方式运行(运行参数中包含-d)，若需要，启动服务并将本进程退出
func CheckDaemon() {
	args := os.Args[1:]

	execArgs := make([]string, 0)
	needDaemon := false
	l := len(args)
	for i := 0; i < l; i++ {
		if strings.Index(args[i], "-d") == 0 || strings.Index(args[i], "--d") == 0 {
			needDaemon = true
			continue
		}
		execArgs = append(execArgs, args[i])
	}

	if !needDaemon { // 无需以服务运行
		return
	}

	proc := exec.Command(os.Args[0], execArgs...)
	err := proc.Start()
	if err != nil {
		panic(err)
	}

	log.Info("PID: ", proc.Process.Pid)
	pidErr := ioutil.WriteFile("./bot.pid", []byte(fmt.Sprintf("%d", proc.Process.Pid)), 0o644)
	if pidErr != nil {
		log.Errorf("save pid file error: %v", pidErr)
	}

	os.Exit(0)
}
