package util

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

// SetupLogs adds hooks to send logs to different destinations depending on level
func SetupLogs(logFileName string) {
	logger := zap.New(zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			MessageKey:  "msg",
			LevelKey:    "level",
			TimeKey:     "ts",
			CallerKey:   "file",
			EncodeLevel: zapcore.CapitalLevelEncoder,
			EncodeTime: func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
				encoder.AppendString(time.Format("2006-01-02 15:04:05"))
			},
			EncodeDuration: func(duration time.Duration, encoder zapcore.PrimitiveArrayEncoder) {
				encoder.AppendInt64(int64(duration) / 1000000)
			},
			EncodeCaller: zapcore.ShortCallerEncoder,
		}), zapcore.AddSync(&lumberjack.Logger{
			Filename:   logFileName, // 日志路径
			MaxSize:    5,           // 日志大小, 单位是M
			MaxAge:     7,           // 最多保存多少天
			MaxBackups: 5,           // 最多备份多少个
			Compress:   true,        // 压缩
		}), zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level >= zapcore.DebugLevel
		})),
		zapcore.NewCore(zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			MessageKey:    "M",
			LevelKey:      "",
			TimeKey:       "",
			NameKey:       "",
			CallerKey:     "",
			StacktraceKey: "",
		}), zapcore.Lock(os.Stdout), zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level >= zapcore.InfoLevel
		})),
	), zap.AddCaller())
	zap.ReplaceGlobals(logger)
}
