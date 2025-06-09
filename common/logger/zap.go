package logger

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"runtime/debug"
	"sync"
)

type ctxKey struct{}

var once sync.Once

var logger *zap.Logger

func Get() *zap.Logger {
	once.Do(
		func() {
			// 输出到文件
			stdout := zapcore.AddSync(os.Stdout)

			// 输出到文件
			file := zapcore.AddSync(&lumberjack.Logger{
				Filename:   "logs/server.log",
				MaxSize:    5,
				MaxBackups: 10,
				MaxAge:     14,
				Compress:   true,
			})

			// 日志等级
			// TODO: 这里后续改为INFO等级
			level := zap.DebugLevel
			levelEnv := os.Getenv("LOG_LEVEL")

			if levelEnv != "" {
				levelFromEnv, err := zapcore.ParseLevel(levelEnv)
				if err != nil {
					log.Println(fmt.Errorf("无效的日志等级，使用默认的INFO等级，Error: %v\n", err))
				}
				level = levelFromEnv
			}
			logLevel := zap.NewAtomicLevelAt(level)

			// 生产环境日志格式
			productionCfg := zap.NewProductionEncoderConfig()
			productionCfg.TimeKey = "timestamp"
			productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

			// 开发环境日志格式
			developmentCfg := zap.NewDevelopmentEncoderConfig()
			developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

			consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
			fileEncoder := zapcore.NewJSONEncoder(productionCfg)

			var gitRevision string

			buildInfo, ok := debug.ReadBuildInfo()
			if ok {
				for _, setting := range buildInfo.Settings {
					if setting.Key == "vcs.revision" {
						gitRevision = setting.Value
						break
					}
				}
			}

			// 创建多输出日志核心，同时输出到控制台和文件
			core := zapcore.NewTee(
				// 控制台输出（开发环境格式）
				zapcore.NewCore(consoleEncoder, stdout, logLevel),
				// 文件输出（生产环境格式）并添加全局字段
				zapcore.NewCore(fileEncoder, file, logLevel).
					With(
						[]zapcore.Field{
							// 记录Git提交版本信息
							zap.String("git_revision", gitRevision),
							// 记录Go版本信息
							zap.String("go_version", buildInfo.GoVersion),
						},
					),
			)
			// 构建最终的Logger实例
			logger = zap.New(core)
		},
	)
	return logger
}

// FromCtx 从上下文中获取日志记录器实例，如果没有则返回全局日志记录器， 如果两者都不存在，则返回一个无操作的日志记录器（nop logger）
func FromCtx(ctx context.Context) *zap.Logger {
	// 尝试从上下文中提取已存在的日志记录器
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return l
	} else if l := logger; l != nil { // 如果上下文中没有日志记录器，则尝试使用全局日志记录器
		return l
	}
	// 如果都没有找到，则返回一个不执行任何操作的日志记录器，避免空指针错误
	return zap.NewNop()
}

// WithCtx 将指定的日志记录器注入到上下文中。
// 如果上下文中已经存在相同的日志记录器，则直接返回原始上下文以避免重复注入。
// 否则，将新的日志记录器附加到上下文并返回新上下文。
func WithCtx(ctx context.Context, l *zap.Logger) context.Context {
	// 从上下文中尝试获取已存在的日志记录器
	if lp, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		// 如果已存在相同的日志记录器，则返回未修改的上下文
		if lp == l {
			return ctx
		}
	}
	// 将新的日志记录器注入上下文并返回
	return context.WithValue(ctx, ctxKey{}, l)
}
