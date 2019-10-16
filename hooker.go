package joker

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"path"
	"time"
)

var (
	defaultErrRollRule = RollRule{
		RotationType: RotationTime,
		Filename:     "default", // 日志文件路径
		MaxSize:      1,         // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups:   10,        // 日志文件最多保存多少个备份
		MaxAge:       7,         // 文件最多保存多少天
		//Compress:   true,      // 是否压缩
	}
	defaultRollRule = RollRule{
		RotationType: RotationTime,

		Filename:   "default_error", // 日志文件路径
		MaxSize:    100,             // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 10,              // 日志文件最多保存多少个备份
		MaxAge:     100,             // 文件最多保存多少天
		//Compress:   true,            // 是否压缩
	}
)

type rotationType int

const (
	RotationTime rotationType = iota
	RotationSize
)

type RollRule struct {
	Logger       string        // 日志模块
	RotationType rotationType  // 日志滚动方式
	Filename     string        // 日志文件名称
	Filepath     string        // 日志文件名称
	MaxSize      int           // 每个日志文件保存的最大尺寸 单位：M
	MaxBackups   int           // 日志文件最多保存多少个备份
	MaxAge       int           // 文件最多保存多少天
	Compress     bool          // 是否压缩
	RotationTime time.Duration //日志切割时间间隔
}

func (rr RollRule) maxAge() time.Duration {
	return time.Duration(rr.MaxAge) * time.Hour * 24
}

func (rr RollRule) fullName() string {
	return path.Join(rr.Filepath, rr.Filename+".log")
}

func getOutHook(outRr *RollRule) zapcore.WriteSyncer {
	return getHook(outRr)

}

func getHook(rr *RollRule) zapcore.WriteSyncer {

	if rr != nil {
		switch rr.RotationType {
		case RotationTime:

			outHook, err := rotatelogs.New(
				rr.fullName()+".%Y%m%d",
				rotatelogs.WithLinkName(rr.fullName()),       // 生成软链，指向最新日志文件
				rotatelogs.WithMaxAge(rr.maxAge()),           // 文件最大保存时间
				rotatelogs.WithRotationTime(rr.RotationTime), // 日志切割时间间隔
			)

			if err != nil {
			} else {
				return zapcore.AddSync(outHook)

			}
		case RotationSize:
			outHook := lumberjack.Logger{
				Filename:   rr.fullName(), // 日志文件路径
				MaxSize:    rr.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
				MaxBackups: rr.MaxBackups, // 日志文件最多保存多少个备份
				MaxAge:     rr.MaxAge,     // 文件最多保存多少天
				Compress:   rr.Compress,   // 是否压缩
			}
			return zapcore.AddSync(&outHook)
		default:
			log.Println("Logging.Hooker.GetHook.RotationType.Error")
		}

	}

	return nil
}

func getErrHook(errRr *RollRule) zapcore.WriteSyncer {
	return getHook(errRr)
}

// 默认全 level 日志切割规则
func GetDefaultRollRule(fName string) *RollRule {
	defaultRollRule.Filename = fName
	return &defaultRollRule
}

// 默认 error 等级日志切割规则
func GetDefaultErrRollRule(fName string) *RollRule {
	defaultErrRollRule.Filename = fName
	return &defaultErrRollRule
}
