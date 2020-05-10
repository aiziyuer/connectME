/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"github.com/aiziyuer/connectDNS/cmd"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

// setupLogs adds hooks to send logs to different destinations depending on level
func setupLogs() {
	logger := zap.New(zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			MessageKey:  "msg",
			LevelKey:    "level",
			TimeKey:     "ts",
			CallerKey:   "file",
			EncodeLevel: zapcore.CapitalLevelEncoder,
			EncodeTime: func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
				encoder.AppendString(time.Format("2020-05-10 00:00:00"))
			},
			EncodeDuration: func(duration time.Duration, encoder zapcore.PrimitiveArrayEncoder) {
				encoder.AppendInt64(int64(duration) / 1000000)
			},
			EncodeCaller: zapcore.ShortCallerEncoder,
		}), zapcore.AddSync(&lumberjack.Logger{
			Filename:   "/var/log/connectDNS/info.log", // 日志路径
			MaxSize:    128,                            // 日志大小, 单位是M
			MaxAge:     7,                              // 最多保存多少天
			MaxBackups: 5,                              // 最多备份多少个
			Compress:   true,                           // 压缩
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

func main() {
	setupLogs()
	cmd.Execute()
}
