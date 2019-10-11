package joker

import (
	"fmt"
	"github.com/braveghost/meteor/errutil"
	"github.com/braveghost/meteor/file"
	"github.com/braveghost/meteor/mode"
	"github.com/braveghost/viper"
	"github.com/micro/go-micro/server"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"log"
	"os"
	"path"
	"time"
)

var (
	// request_id key
	requestIdKey = "request_id"
	// 环境变量
	envKeyLogPathPrefix = "LOGGING_LOGGER_PATH"
	// 默认日志存放路径件相对路径
	logPathPrefix = "log"

	defaultLogger       *zap.SugaredLogger
	defaultLoggerName   = "init_default"
	defaultLoggerStatus bool

	defaultEncoderConfig = &zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器
	}

	// 从 viper 获取日志配置信息
	viperKeyLoggerPath = "service.log_path"
	viperKeyLoggerName = "service.log_name"
)

var (
	LogPathErr = errors.New("log path error")
	LogNameErr = errors.New("log name error")
)

func init() {
	// 自动设置当前项目路径为日志路径
	SetLogPathAuto()
	GetLogger(defaultLoggerName, mode.ModeLocal)

}

func SetRequestIdKey(key string) {
	requestIdKey = key
}

// 根据环境变量设置日志文件存放路径
func SetLogPathByEnv() {
	pt := os.Getenv(envKeyLogPathPrefix)
	if pt != "" {
		logPathPrefix = path.Join(pt, logPathPrefix)
	}
	if !file.DirNotExistCreate(logPathPrefix) {
		log.Panicf("Logging.SetLogPathAuto.DirNotExistCreate.Error || path=%v | err=%v", logPathPrefix, LogPathErr)
	}

}

// 根据 viper 配置信息存储路径
func SetLogPathByViper(key string) {
	SetLogPath(viper.GetString(key))
	if !file.DirNotExistCreate(logPathPrefix) {
		log.Panicf("Logging.SetLogPathByViper.DirNotExistCreate.Error | path=%v | err=%v", logPathPrefix, LogPathErr)
	}
}

// 默认设置日志文件存放路径
func SetLogPathAuto() {
	pt, _ := os.Getwd()
	SetLogPath(path.Join(pt, logPathPrefix))
}

// 手动设置固定路径
func SetLogPath(pt string) {
	logPathPrefix = pt
	log.Printf("Logging.SetLogPath || path=%v", logPathPrefix)
}

func Fields(fields ...zap.Field) []zap.Field {
	return fields
}

// 根据 mode 获取日志输出等级
func getLevel(md mode.ModeType) zapcore.Level {
	level := zapcore.DebugLevel
	if md == mode.ModePro {
		level = zapcore.InfoLevel
	}
	return level
}

// 初始化 logger
func InitLogger(md mode.ModeType, outRr, errRr *RollRule, fields []zap.Field, ec *zapcore.EncoderConfig) *zap.SugaredLogger {

	if ec == nil {
		// 兜底配置
		ec = defaultEncoderConfig
	}

	if outRr == nil {
		// 兜底配置
		outRr = &defaultRollRule
	}

	// 设置日志级别
	level := getLevel(md)

	var outWriter []zapcore.WriteSyncer

	outHook := getOutHook(outRr)
	if outHook != nil {
		outWriter = append(outWriter, outHook)
	}

	if level == zapcore.DebugLevel {
		// 打印到控制台和文件
		outWriter = append(outWriter, zapcore.AddSync(os.Stdout))
	}
	outputCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(*ec), // 编码器配置
		zapcore.NewMultiWriteSyncer(outWriter...),
		zap.NewAtomicLevelAt(level), // 日志级别
	)
	fmt.Println(outputCore)
	var options = []zap.Option{
		// 开启堆栈跟踪
		zap.AddCaller(),
		// 因为 operate 包装了一层所以堆栈信息加1
		zap.AddCallerSkip(1),
		// 开启文件及行号
		zap.Development(),
		// 设置初始化字段
		zap.Fields(fields...),
		zap.ErrorOutput(zapcore.AddSync(os.Stderr)),
	}

	var errCore zapcore.Core

	if errRr != nil {
		// 无默认, 错误日志规则传入 nil 表示不独立写错误日志文件

		var errWriter []zapcore.WriteSyncer

		errHook := getErrHook(errRr)
		if errHook != nil {
			errWriter = append(errWriter, errHook)
		}

		if level == zapcore.DebugLevel {
			// 打印到控制台和文件
			errWriter = append(errWriter, zapcore.AddSync(os.Stdout))
		}
		errCore = zapcore.NewCore(
			zapcore.NewConsoleEncoder(*ec), // 编码器配置
			zapcore.NewMultiWriteSyncer(errWriter...),
			zap.NewAtomicLevelAt(zap.ErrorLevel), // 日志级别
		)
	}

	cores := zapcore.NewTee(
		outputCore,
		errCore,
	)

	// 构造日志
	logger := zap.New(cores, options...)
	return logger.Sugar()
}

// 日志中间件
func NewLogWrap() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {

			Infow("[middle start]", "Service", req.Service(), "Method", req.Method(), "req", req)
			s := time.Now()

			var err error
			err = fn(ctx, req, rsp)
			defer Infow("[middle end]", "Service", req.Service(), "Method", req.Method(), "time", time.Since(s), "req", req, "rsp", rsp, "err", err)
			return err
		}
	}
}

// 初始化日志
func GetLogger(name string, md mode.ModeType) *zap.SugaredLogger {

	// 确定路径
	if !file.IsDir(logPathPrefix) {
		if name == defaultLoggerName {
			log.Printf("GetLogger.VerifyLogPath.Error || path=%v | name=%v |err=%v\n",
				logPathPrefix, name, LogPathErr)
		}
	} else {
		initDefaultLogger(name, md)
	}

	return defaultLogger
}

func initDefaultLogger(name string, md mode.ModeType) {
	defaultLogger = InitLogger(
		md,
		GetDefaultRollRule(name),
		GetDefaultErrRollRule(name+"_error"),
		Fields(zap.String("srv_name", name)),
		nil,
	)
	if defaultLogger != nil {
		defaultLoggerStatus = true
	}
}

// 修改 viper 日志配置信息 key
func SetViperKey(name, value string) {
	switch name {
	case "name":
		viperKeyLoggerName = value
	case "path":
		viperKeyLoggerPath = value
	}
}

// 通过 viper 设置日志路径并初始化 logger
func GetLoggerByViper(md string) *zap.SugaredLogger {
	name := viper.GetString(viperKeyLoggerName)
	fmt.Println(name)
	if name == "" {
		errutil.CheckErrPanic(LogNameErr)
	}

	m := mode.ModeType(md)
	SetLogPathByViper(viperKeyLoggerPath)
	return GetLogger(name, m)
}

// 从 context 获取request id
func GetRequestId(ctx context.Context) interface{} {
	return ctx.Value(requestIdKey)
}
