package joker

import (
	"github.com/braveghost/meteor/file"
	"github.com/braveghost/meteor/mode"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"log"
	"os"
	"path"
	"sync"
)

const (
	// 环境变量
	envKeyLogPath = "LOGGING_JOKER_PATH"
)

var (
	// initflag

	initFlag bool

	// traceId key
	traceIdKey = "trace_id"

	// 默认日志存放路径件相对路径
	defaultLoggerPath     = "log"
	defaultLoggerFileName = "joker"
	defaultServiceName    string

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

	loggers = map[string]*Logging{}

	defaultLogger *Logging
	traceOnce     sync.Once
	pathOnce      sync.Once
)

var (
	LogPathError    = errors.New("log path error")
	LoggerInitError = errors.New("logger init error")
)

func init() {
	// 自动设置当前项目路径为日志路径
	InitLogger(mode.ModeLocal)
}

func SetServiceName(name string) {
	defaultServiceName = name
}

func SetLogName(name string) {
	defaultLoggerFileName = name
}

func SetTraceIdKey(key string) {
	traceOnce.Do(func() {
		traceIdKey = key
	})

}

// 根据环境变量设置日志文件存放路径
func SetLogPathByEnv() {
	pt := os.Getenv(envKeyLogPath)
	if pt != "" {
		SetLogPath(path.Join(pt, defaultLoggerPath))
	}

}

// 默认设置日志文件存放路径
func SetLogPathAuto() {
	pt, _ := os.Getwd()
	SetLogPath(path.Join(pt, defaultLoggerPath))

}

const setLogPathLogMsg = "Logging.SetLogPathAuto.DirNotExistCreate.Error || path=%s | err=%s"

// 手动设置固定路径
func SetLogPath(pt string) {
	pathOnce.Do(func() {
		defaultLoggerPath = pt
		if !file.DirNotExistCreate(defaultLoggerPath) {
			log.Panicf(setLogPathLogMsg, defaultLoggerPath, LogPathError.Error())
		}

	})
}

// 获取默认日志对象
func GetLogger(name string) *Logging {
	if lg, ok := loggers[name]; ok {
		return lg
	}
	return nil
}

const newLoggerPathLogMsg = "GetLogger.VerifyLogPath.Error || path=%s | name=%s |err=%s\n"

// 初始化日志对象对象
func NewLogger(conf *LoggingConf) error {
	// 确定路径
	if !file.IsDir(conf.Path) {
		log.Printf(newLoggerPathLogMsg, conf.Path, conf.Name, LogPathError.Error())
		return LogPathError
	}
	tmp := &Logging{
		conf: conf,
	}
	tmp.initLogger()
	if !tmp.status {
		return LoggerInitError
	}
	loggers[conf.Name] = tmp
	return nil
}

// 初始化 default logger
func InitLogger(md mode.ModeType) {
	defaultLogger = &Logging{
		conf: &LoggingConf{
			ServiceName: defaultServiceName,
			Name:        defaultLoggerFileName,
			Path:        defaultLoggerPath,
			Mode:        md,
			OutRr:       GetDefaultRollRule(defaultLoggerFileName),
			ErrRr:       GetDefaultErrRollRule(defaultLoggerFileName + "_error"),
		},
	}
	defaultLogger.initLogger()
}

// 从 context 获取request id
func GetTraceId(ctx context.Context) interface{} {
	return ctx.Value(traceIdKey)
}

type LoggingConf struct {
	*zapcore.EncoderConfig
	ServiceName  string
	Name         string
	Path         string
	Mode         mode.ModeType
	OutRr, ErrRr *RollRule
	Fields       []zap.Field // 扩展输出字段
}

func (lc LoggingConf) GetPath() string {
	if len(lc.Path) == 0 {
		return defaultLoggerPath
	}
	return lc.Path
}

func (lc LoggingConf) GetName() string {
	if len(lc.Name) == 0 {
		return defaultLoggerFileName
	}
	return lc.Name
}

func (lc LoggingConf) GetErrorName() string {
	if len(lc.Name) == 0 {
		return defaultLoggerFileName + "_error"
	}
	return lc.Name + "_error"
}

func (lc LoggingConf) ExtendField() []zap.Field {
	if len(lc.ServiceName) == 0 {
		return lc.Fields
	}
	return append(lc.Fields, zap.String("service_name", lc.ServiceName))
}

type Logging struct {
	logger *zap.SugaredLogger
	status bool
	conf   *LoggingConf
	level  zapcore.Level
}

// 根据 mode 设置日志输出等级
func (lg *Logging) setLevel() {
	lg.level = zapcore.DebugLevel
	if lg.conf.Mode == mode.ModePro {
		lg.level = zapcore.InfoLevel
	}
}

// 初始化 logger
func (lg *Logging) initLogger() {
	lg.setLevel()
	encoderConfig := lg.conf.EncoderConfig
	if encoderConfig == nil {
		// 兜底配置
		encoderConfig = defaultEncoderConfig
	}
	var (
		cores   = []zapcore.Core{}
		errCore zapcore.Core
	)

	cores = append(cores, lg.getOutputCore(encoderConfig))

	errCore = lg.getErrorCore(encoderConfig)
	if errCore != nil {
		cores = append(cores, errCore)
	}

	tee := zapcore.NewTee(cores...)

	var options = []zap.Option{
		// 开启堆栈跟踪
		zap.AddCaller(),
		// 因为 operate 包装了一层所以堆栈信息加1
		zap.AddCallerSkip(2),
		// 开启文件及行号
		zap.Development(),
		// 设置初始化字段
		zap.Fields(lg.conf.ExtendField()...),
		zap.ErrorOutput(zapcore.AddSync(os.Stderr)),
	}
	// 构造日志
	lg.logger = zap.New(tee, options...).Sugar()
	lg.status = true
}

func (lg *Logging) getOutputCore(encoderConfig *zapcore.EncoderConfig) zapcore.Core {

	outRr := lg.conf.OutRr

	if outRr == nil {
		// 兜底配置
		outRr = &defaultRollRule

	}

	outRr.Filepath = lg.conf.GetPath()
	outRr.Filename = lg.conf.GetName()
	// 设置日志级别

	var (
		outHook   zapcore.WriteSyncer
		outWriter []zapcore.WriteSyncer
	)
	if initFlag {
		outHook = getOutHook(outRr)
		if outHook != nil {
			outWriter = append(outWriter, outHook)
		}
	}

	if lg.level == zapcore.DebugLevel {
		// 打印到控制台和文件
		outWriter = append(outWriter, zapcore.AddSync(os.Stdout))
	}
	return zapcore.NewCore(
		zapcore.NewConsoleEncoder(*encoderConfig), // 编码器配置
		zapcore.NewMultiWriteSyncer(outWriter...),
		zap.NewAtomicLevelAt(lg.level), // 日志级别
	)
}

// 生成错误日志引擎
func (lg *Logging) getErrorCore(encoderConfig *zapcore.EncoderConfig) zapcore.Core {
	errRr := lg.conf.ErrRr

	if errRr != nil {
		// 无默认, 错误日志规则传入 nil 表示不独立写错误日志文件
		errRr.Filepath = lg.conf.GetPath()
		errRr.Filename = lg.conf.GetErrorName()

		var (
			errWriter []zapcore.WriteSyncer
			errHook   zapcore.WriteSyncer
		)
		if initFlag {
			errHook = getErrHook(errRr)

			if errHook != nil {
				errWriter = append(errWriter, errHook)
			}

		}
		if lg.level == zapcore.DebugLevel {
			// 打印到控制台和文件
			errWriter = append(errWriter, zapcore.AddSync(os.Stdout))
		}
		return zapcore.NewCore(
			zapcore.NewConsoleEncoder(*encoderConfig), // 编码器配置
			zapcore.NewMultiWriteSyncer(errWriter...),
			zap.NewAtomicLevelAt(zap.ErrorLevel), // 日志级别
		)
	}
	return nil
}

func (lg *Logging) FullPath() string {
	return path.Join(lg.conf.Path, lg.conf.Name)
}

// Debug uses fmt.Sprint to construct and log a message.
func (lg *Logging) Debug(args ...interface{}) {
	if lg.status {
		lg.logger.Debug(args...)
	}
}

// Info uses fmt.Sprint to construct and log a message.
func (lg *Logging) Info(args ...interface{}) {
	if lg.status {
		lg.logger.Info(args...)
	}
}

// Warn uses fmt.Sprint to construct and log a message.
func (lg *Logging) Warn(args ...interface{}) {
	if lg.status {
		lg.logger.Warn(args...)
	}
}

// Error uses fmt.Sprint to construct and log a message.
func (lg *Logging) Error(args ...interface{}) {
	if lg.status {
		lg.logger.Error(args...)
	}
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (lg *Logging) DPanic(args ...interface{}) {
	if lg.status {
		lg.logger.DPanic(args...)
	}
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (lg *Logging) Panic(args ...interface{}) {
	if lg.status {
		lg.logger.Panic(args...)
	}
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (lg *Logging) Fatal(args ...interface{}) {
	if lg.status {
		lg.logger.Fatal(args...)
	}
}

// Debugf uses fmt.Sprintf to log a templated message.
func (lg *Logging) Debugf(template string, args ...interface{}) {
	if lg.status {
		lg.logger.Debugf(template, args...)
	}
}

// Infof uses fmt.Sprintf to log a templated message.
func (lg *Logging) Infof(template string, args ...interface{}) {
	if lg.status {
		lg.logger.Infof(template, args...)
	}
}

// Warnf uses fmt.Sprintf to log a templated message.
func (lg *Logging) Warnf(template string, args ...interface{}) {
	if lg.status {
		lg.logger.Warnf(template, args...)
	}
}

// Errorf uses fmt.Sprintf to log a templated message.
func (lg *Logging) Errorf(template string, args ...interface{}) {
	if lg.status {
		lg.logger.Errorf(template, args...)
	}
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (lg *Logging) DPanicf(template string, args ...interface{}) {
	if lg.status {
		lg.logger.DPanicf(template, args...)
	}
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (lg *Logging) Panicf(template string, args ...interface{}) {
	if lg.status {
		lg.logger.Panicf(template, args...)
	}
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (lg *Logging) Fatalf(template string, args ...interface{}) {
	if lg.status {
		lg.logger.Fatalf(template, args...)
	}
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func (lg *Logging) Debugw(msg string, keysAndValues ...interface{}) {
	if lg.status {
		lg.logger.Debugw(msg, keysAndValues...)
	}
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) Infow(msg string, keysAndValues ...interface{}) {
	if lg.status {
		lg.logger.Infow(msg, keysAndValues...)
	}
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) Warnw(msg string, keysAndValues ...interface{}) {
	if lg.status {
		lg.logger.Warnw(msg, keysAndValues...)
	}
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) Errorw(msg string, keysAndValues ...interface{}) {
	if lg.status {
		lg.logger.Errorw(msg, keysAndValues...)
	}
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) DPanicw(msg string, keysAndValues ...interface{}) {
	if lg.status {
		lg.logger.DPanicw(msg, keysAndValues...)
	}
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func (lg *Logging) Panicw(msg string, keysAndValues ...interface{}) {
	if lg.status {
		lg.logger.Panicw(msg, keysAndValues...)
	}
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func (lg *Logging) Fatalw(msg string, keysAndValues ...interface{}) {
	if lg.status {
		lg.logger.Fatalw(msg, keysAndValues...)
	}
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func (lg *Logging) Debugwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.status {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.Debugw(msg, keysAndValues..., )
	}
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) Infowc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.status {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.Infow(msg, keysAndValues...)
	}
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) Warnwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.status {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.Warnw(msg, keysAndValues...)
	}
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) Errorwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.status {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.Errorw(msg, keysAndValues...)
	}
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) DPanicwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.status {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.DPanicw(msg, keysAndValues...)
	}
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func (lg *Logging) Panicwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.status {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.Panicw(msg, keysAndValues...)
	}
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func (lg *Logging) Fatalwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.status {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.Fatalw(msg, keysAndValues...)
	}
}

func (lg *Logging) Sync() {
	if lg.status {
		lg.logger.Sync()
	}
}
