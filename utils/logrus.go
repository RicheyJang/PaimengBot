package utils

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

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
