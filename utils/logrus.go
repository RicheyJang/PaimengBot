package utils

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/utils/consts"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type SimpleFormatter struct{}

const stringOfSymbol = "[bot]"
const stringOfStarter = ": "
const stringOfIgnoreTip = "This is a log generated from the forced abort, please ignore!"

var stringsOfBase64End = []string{`"`, "]", "(id"}

func (f SimpleFormatter) Format(entry *log.Entry) ([]byte, error) {
	var output bytes.Buffer
	formatEntryMessage(entry)
	// 标识
	output.WriteString(stringOfSymbol)
	// 时间
	output.WriteString(entry.Time.Format("[2006-01-02 15:04:05.000ms]"))
	// 等级
	output.WriteRune('[')
	output.WriteString(entry.Level.String())
	output.WriteRune(']')
	// 消息
	output.WriteString(stringOfStarter)
	output.WriteString(entry.Message)
	// 键值对
	output.WriteRune(' ')
	for k, val := range entry.Data {
		output.WriteString(k)
		output.WriteRune(':')
		output.WriteString(cast.ToString(val))
		output.WriteRune(' ')
	}
	output.WriteRune('\n')
	return output.Bytes(), nil
}

func formatEntryMessage(entry *log.Entry) {
	// 将Hook导致的Panic->log.Error的日志等级降为Debug并提示忽略
	if entry.Level <= log.ErrorLevel && strings.Contains(entry.Message, consts.AbortLogIgnoreSymbol) {
		entry.Level = log.DebugLevel
		entry.Message = stringOfIgnoreTip
	}
	// Info 消息过长，尝试缩减
	if entry.Level == log.InfoLevel && len(entry.Message) >= 500 {
		// Base64 缩减
		base64Index := strings.Index(entry.Message, "base64://")
		if base64Index != -1 && base64Index+20 < len(entry.Message) { // 含Base64内容
			lastIndex := getEntryMsgBase64LastIndex(entry, base64Index)
			if lastIndex < base64Index+20 {
				lastIndex = len(entry.Message) - 5
			}
			// 截断
			if lastIndex > base64Index+20 {
				entry.Message = entry.Message[:base64Index+20] + "...... " + entry.Message[lastIndex:]
			} else {
				entry.Message = entry.Message[:base64Index+20] + "......"
			}
		}
		// 其它缩减
	}
}

func getEntryMsgBase64LastIndex(entry *log.Entry, base64Index int) (lastIndex int) {
	for _, substr := range stringsOfBase64End {
		lastIndex = strings.LastIndex(entry.Message, substr)
		if lastIndex > base64Index {
			return
		}
	}
	return -1
}

type loggerGorm struct {
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
}

func NewGormLogger() *loggerGorm {
	return &loggerGorm{
		SkipErrRecordNotFound: true,
	}
}

func (l *loggerGorm) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *loggerGorm) Info(ctx context.Context, s string, args ...interface{}) {
	log.WithContext(ctx).Infof(s, args)
}

func (l *loggerGorm) Warn(ctx context.Context, s string, args ...interface{}) {
	log.WithContext(ctx).Warnf(s, args)
}

func (l *loggerGorm) Error(ctx context.Context, s string, args ...interface{}) {
	log.WithContext(ctx).Errorf(s, args)
}

func (l *loggerGorm) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := log.Fields{}
	if l.SourceField != "" {
		fields[l.SourceField] = utils.FileWithLineNum()
	}
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		fields[log.ErrorKey] = err
		log.WithContext(ctx).WithFields(fields).Errorf("%s [%s]", sql, elapsed)
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		log.WithContext(ctx).WithFields(fields).Warnf("%s [%s]", sql, elapsed)
		return
	}

	log.WithContext(ctx).WithFields(fields).Debugf("%s [%s]", sql, elapsed)
}

type LoggerCron struct{}

func NewCronLogger() *LoggerCron {
	return new(LoggerCron)
}

func (l *LoggerCron) Info(msg string, keysAndValues ...interface{}) {
	if msg == "wake" || msg == "run" {
		return
	}
	log.Info("cron msg: ", msg)
}

func (l *LoggerCron) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Error("cron msg: ", msg, "err: ", err)
}
