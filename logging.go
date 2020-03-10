package logging

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
	"time"
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
	defaultLoggerPath     = ""
	defaultLoggerFileName = "joker"
	defaultServiceName    string

	defaultEncoderConfig = &zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // 小写编码器
		EncodeTime:     MilliSecondTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	loggers = map[string]*Logging{}

	skip          = 1
	defaultLogger *Logging
	traceOnce     sync.Once
	pathOnce      sync.Once
)

var (
	LogPathError    = errors.New("log path error")
	LoggerInitError = errors.New("logger init error")
)

func MilliSecondTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.0000"))
}

func init() {
	// 自动设置当前项目路径为日志路径
	InitLogger(mode.ModeLocal)
	initFlag = true
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
		if len(pt) == 0 {
			return
		}
		defaultLoggerPath = pt
		if !file.DirNotExistCreate(defaultLoggerPath) {
			log.Panicf(setLogPathLogMsg, defaultLoggerPath, LogPathError.Error())
		}

	})
}

// 获取默认日志对象
func Logger(name string) *Logging {
	if lg, ok := loggers[name]; ok {
		return lg
	}
	return &Logging{}
}

const newLoggerPathLogMsg = "GetLogger.VerifyLogPath.Error || path=%s | name=%s |err=%s\n"

// 初始化日志对象对象
func NewLogger(conf *Options) error {
	// 确定路径
	//if !file.IsDir(conf.Path) {
	//	log.Printf(newLoggerPathLogMsg, conf.Path, conf.FileName, LogPathError.Error())
	//	return LogPathError
	//}
	tmp := &Logging{
		opts: conf,
	}
	tmp.initLogger()
	if !tmp.status {
		return LoggerInitError
	}
	loggers[conf.FileName] = tmp
	return nil
}

// 初始化 default logger
func InitLogger(md mode.ModeType) {
	skip = 2
	defaultLogger = &Logging{
		opts: &Options{
			ServiceName: defaultServiceName,
			FileName:    defaultLoggerFileName,
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

type encoderOption func(*zapcore.EncoderConfig)

var (
	_FlagOpenColor bool
	_FlagLowercase bool
)

func OpenColor() {
	_FlagOpenColor = true
}

type timeLayout string

const (
	TimeLayoutNano   timeLayout = "2006-01-02 15:04:05.9999999"
	TimeLayoutMicro  timeLayout = "2006-01-02 15:04:05.999"
	TimeLayoutSecond timeLayout = "2006-01-02 15:04:05"

	TimeLayoutDaily    timeLayout = "20060102"
	TimeLayoutHourly   timeLayout = "2006010215"
	TimeLayoutSecondly timeLayout = "200601021505"
)

func (tl timeLayout) String() string {
	return string(tl)
}

func TimeFormater(layout timeLayout) encoderOption {
	return func(c *zapcore.EncoderConfig) {
		c.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(layout.String()))
		}
	}
}

type optionRollRules struct {
	All     *RollRule
	Debug   *RollRule
	Info    *RollRule
	Warning *RollRule
	Error   *RollRule
	Fatal   *RollRule
}

type Options struct {
	*zapcore.EncoderConfig
	ServiceName string
	FileName    string
	Path        string
	Mode        mode.ModeType
	//RollRules   *optionRollRules
	OutRr  *RollRule
	ErrRr  *RollRule
	Fields []zap.Field // 扩展输出字段

	encoder   []encoderOption
	OpenColor bool
	Lowercase bool
}

func (lc Options) GetPath() string {
	if len(lc.Path) == 0 {
		return defaultLoggerPath
	}
	return lc.Path
}

func (lc Options) GetName() string {
	if len(lc.FileName) == 0 {
		return defaultLoggerFileName
	}
	return lc.FileName
}

func (lc Options) GetErrorName() string {
	if len(lc.FileName) == 0 {
		return defaultLoggerFileName + "_error"
	}
	return lc.FileName + "_error"
}

func (lc Options) ExtendField() []zap.Field {
	if len(lc.ServiceName) == 0 {
		return lc.Fields
	}
	return append(lc.Fields, zap.String("service_name", lc.ServiceName))
}

type Logging struct {
	logger *zap.SugaredLogger
	status bool
	opts   *Options
	level  zapcore.Level
}

// 根据 mode 设置日志输出等级
func (lg *Logging) setLevel() {
	lg.level = zapcore.DebugLevel
	if lg.opts.Mode == mode.ModePro {
		lg.level = zapcore.InfoLevel
	}
}

// 初始化 logger
func (lg *Logging) initLogger() {
	lg.setLevel()
	encoderConfig := lg.opts.EncoderConfig
	if encoderConfig == nil {
		// 兜底配置
		encoderConfig = defaultEncoderConfig
	}

	if _FlagOpenColor {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
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

	// 构造日志
	lg.logger = zap.New(tee).WithOptions( // 开启堆栈跟踪
		zap.AddCaller(),
		// 因为 operate 包装了一层所以堆栈信息加1
		zap.AddCallerSkip(skip),
		// 开启文件及行号
		//zap.Development(),
		// 设置初始化字段
		zap.Fields(lg.opts.ExtendField()...),
		zap.ErrorOutput(zapcore.AddSync(os.Stderr)) ).Sugar()
	lg.status = true
}

func (lg *Logging) getOutputCore(encoderConfig *zapcore.EncoderConfig) zapcore.Core {

	outRr := lg.opts.OutRr

	if outRr == nil {
		// 兜底配置
		outRr = &defaultRollRule

	}

	outRr.Filepath = lg.opts.GetPath()
	outRr.Filename = lg.opts.GetName()
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
	errRr := lg.opts.ErrRr

	if errRr != nil {
		// 无默认, 错误日志规则传入 nil 表示不独立写错误日志文件
		errRr.Filepath = lg.opts.GetPath()
		if len(errRr.Filepath) == 0 {
			return nil
		}
		errRr.Filename = lg.opts.GetErrorName()

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
	return path.Join(lg.opts.Path, lg.opts.FileName)
}
func (lg *Logging) loggerStatus(args ...interface{}) bool {
	status := lg.status
	if !status {
		args = append([]interface{}{"GetLoggerIsNull"}, args...)
		log.Println(args...)
	}
	return status
}
func (lg *Logging) loggerStatusMsg(msg string, args ...interface{}) bool {
	status := lg.status
	if !status {
		args = append([]interface{}{"GetLoggerIsNull", msg}, args...)
		log.Println(args...)
	}
	return status
}
func (lg *Logging) loggerStatusFormat(format string, args ...interface{}) bool {
	status := lg.status
	if !status {
		args = append([]interface{}{"GetLoggerIsNull"}, args...)
		log.Printf(format, args...)
	}
	return status
}

// Debug uses fmt.Sprint to construct and log a message.
func (lg *Logging) Debug(args ...interface{}) {
	if lg.loggerStatus(args...) {
		lg.logger.Debug(args...)
	}
}

// Info uses fmt.Sprint to construct and log a message.
func (lg *Logging) Info(args ...interface{}) {
	if lg.loggerStatus(args...) {
		lg.logger.Info(args...)
	}
}

// Warn uses fmt.Sprint to construct and log a message.
func (lg *Logging) Warn(args ...interface{}) {
	if lg.loggerStatus(args...) {
		lg.logger.Warn(args...)
	}
}

// Error uses fmt.Sprint to construct and log a message.
func (lg *Logging) Error(args ...interface{}) {
	if lg.loggerStatus(args...) {
		lg.logger.Error(args...)
	}
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (lg *Logging) DPanic(args ...interface{}) {
	if lg.loggerStatus(args...) {
		lg.logger.DPanic(args...)
	}
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (lg *Logging) Panic(args ...interface{}) {
	if lg.loggerStatus(args...) {
		lg.logger.Panic(args...)
	}
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (lg *Logging) Fatal(args ...interface{}) {
	if lg.loggerStatus(args...) {
		lg.logger.Fatal(args...)
	}
}

// Debugf uses fmt.Sprintf to log a templated message.
func (lg *Logging) Debugf(template string, args ...interface{}) {
	if lg.loggerStatusFormat(template, args...) {
		lg.logger.Debugf(template, args...)
	}
}

// Infof uses fmt.Sprintf to log a templated message.
func (lg *Logging) Infof(template string, args ...interface{}) {
	if lg.loggerStatusFormat(template, args...) {
		lg.logger.Infof(template, args...)
	}
}

// Warnf uses fmt.Sprintf to log a templated message.
func (lg *Logging) Warnf(template string, args ...interface{}) {
	if lg.loggerStatusFormat(template, args...) {
		lg.logger.Warnf(template, args...)
	}
}

// Errorf uses fmt.Sprintf to log a templated message.
func (lg *Logging) Errorf(template string, args ...interface{}) {
	if lg.loggerStatusFormat(template, args...) {
		lg.logger.Errorf(template, args...)
	}
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (lg *Logging) DPanicf(template string, args ...interface{}) {
	if lg.loggerStatusFormat(template, args...) {
		lg.logger.DPanicf(template, args...)
	}
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (lg *Logging) Panicf(template string, args ...interface{}) {
	if lg.loggerStatusFormat(template, args...) {
		lg.logger.Panicf(template, args...)
	}
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (lg *Logging) Fatalf(template string, args ...interface{}) {
	if lg.loggerStatusFormat(template, args...) {
		lg.logger.Fatalf(template, args...)
	}
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func (lg *Logging) Debugw(msg string, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		lg.logger.Debugw(msg, keysAndValues...)
	}
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) Infow(msg string, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		lg.logger.Infow(msg, keysAndValues...)
	}
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) Warnw(msg string, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		lg.logger.Warnw(msg, keysAndValues...)
	}
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) Errorw(msg string, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		lg.logger.Errorw(msg, keysAndValues...)
	}
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) DPanicw(msg string, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		lg.logger.DPanicw(msg, keysAndValues...)
	}
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func (lg *Logging) Panicw(msg string, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		lg.logger.Panicw(msg, keysAndValues...)
	}
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func (lg *Logging) Fatalw(msg string, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		lg.logger.Fatalw(msg, keysAndValues...)
	}
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func (lg *Logging) Debugwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.Debugw(msg, keysAndValues..., )
	}
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) Infowc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.Infow(msg, keysAndValues...)
	}
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) Warnwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.Warnw(msg, keysAndValues...)
	}
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) Errorwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.Errorw(msg, keysAndValues...)
	}
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func (lg *Logging) DPanicwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.DPanicw(msg, keysAndValues...)
	}
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func (lg *Logging) Panicwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.Panicw(msg, keysAndValues...)
	}
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func (lg *Logging) Fatalwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if lg.loggerStatusMsg(msg, keysAndValues...) {
		keysAndValues = append(keysAndValues, traceIdKey, GetTraceId(ctx))
		lg.logger.Fatalw(msg, keysAndValues...)
	}
}

func (lg *Logging) Sync() {
	if lg.status {
		lg.logger.Sync()
	}
}
