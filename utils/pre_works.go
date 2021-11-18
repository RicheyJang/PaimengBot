package utils

import (
	"io"
	"os"
	"path/filepath"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/fsnotify/fsnotify"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func init() {
	pflag.StringP("server", "s", "ws://127.0.0.1:6700/", "the websocket server address")
	pflag.StringSliceP("superuser", "u", []string{}, "all superusers' id")
	pflag.StringP("nickname", "n", "派蒙", "the bot's nickname")
	pflag.StringP("log", "l", "info", "the level of logging")
	pflag.Parse()
}

// DoPreWorks 进行全局初始化工作
func DoPreWorks() {
	// 读取主配置
	viper.SetDefault("logDate", 30)
	err := flushMainConfig(".", "config-main.yaml")
	if err != nil {
		log.Fatal("FlushMainConfig err: ", err)
		return
	}
	err = setupLogger()
	if err != nil {
		log.Fatal("setupLogger err: ", err)
		return
	}
}

// 设置日志
func setupLogger() error {
	// 日志等级
	log.SetLevel(log.InfoLevel)
	if l, ok := flagLToLevel[viper.GetString("log")]; ok {
		log.SetLevel(l)
	}
	// 日志格式
	log.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[bot][%time%][%lvl%]: %msg% \n",
	})
	// 日志滚动切割
	logf, err := rotatelogs.New(
		"./log/bot-%Y-%m-%d.log",
		rotatelogs.WithLinkName("./log/bot.log"),
		rotatelogs.WithMaxAge(time.Duration(viper.GetInt("logDate"))*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Error("Get rotate logs err: ", err)
		return err
	}
	// 日志输出
	logWriter := io.MultiWriter(os.Stdout, logf)
	log.SetOutput(logWriter) // logrus 设置日志的输出方式
	return nil
}

var flagLToLevel = map[string]log.Level{
	"debug": log.DebugLevel,
	"Debug": log.DebugLevel,
	"DEBUG": log.DebugLevel,
	"info":  log.InfoLevel,
	"Info":  log.InfoLevel,
	"INFO":  log.InfoLevel,
	"warn":  log.WarnLevel,
	"Warn":  log.WarnLevel,
	"WARN":  log.WarnLevel,
	"error": log.ErrorLevel,
	"Error": log.ErrorLevel,
	"ERROR": log.ErrorLevel,
}

// 从文件和命令行中刷新所有主配置，若文件不存在将会把配置写入该文件
func flushMainConfig(configPath string, configFileName string) error {
	// 从命令行读取
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Error("FlushMainConfig error in BindPFlags")
		return err
	}
	// 从文件读取
	viper.AddConfigPath(configPath)
	viper.SetConfigFile(configFileName)
	fullPath := filepath.Join(configPath, configFileName)
	//fileType := filepath.Ext(fullPath)
	//viper.SetConfigType(fileType)
	if FileExists(fullPath) { // 配置文件已存在：读出配置
		err = viper.ReadInConfig()
		if err != nil {
			log.Error("FlushMainConfig error in ReadInConfig")
			return err
		}
	} else { // 配置文件不存在：写入配置
		err = viper.SafeWriteConfigAs(fullPath)
		if err != nil {
			log.Error("FlushMainConfig error in SafeWriteConfig")
			return err
		}
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
