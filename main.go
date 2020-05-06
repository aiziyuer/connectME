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
	"fmt"
	"github.com/aiziyuer/connectDNS/cmd"
	"github.com/gogf/gf/os/gfile"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type WriterHook struct {
	Writer    io.Writer
	LogLevels []log.Level
}

func (hook *WriterHook) Fire(entry *log.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

func (hook *WriterHook) Levels() []log.Level {
	return hook.LogLevels
}

// setupLogs adds hooks to send logs to different destinations depending on level
func setupLogs() {
	log.SetOutput(ioutil.Discard) // Send all logs to nowhere by default

	log.AddHook(&WriterHook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})

	log.AddHook(&WriterHook{ // Send logs with level higher than warning to stderr
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
		},
	})

	logDir := "/var/log/connectDNS"
	_ = gfile.Mkdir(gfile.Dir(logDir))
	traceLogFile, err := os.OpenFile(path.Join(logDir, "trace.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("log file(%s) open failed: %s", path.Join(logDir, "trace.log"), err))
	}
	log.AddHook(&WriterHook{ // Send info and debug logs to stdout
		Writer: traceLogFile,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
			log.InfoLevel,
			log.DebugLevel,
			log.TraceLevel,
		},
	})

	warnLogFile, err := os.OpenFile(path.Join(logDir, "warn.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("log file(%s) open failed: %s", path.Join(logDir, "warn.log"), err))
	}
	log.AddHook(&WriterHook{ // Send info and debug logs to stdout
		Writer: warnLogFile,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})

	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   false,
	})
}

func main() {
	setupLogs()
	cmd.Execute()
}
