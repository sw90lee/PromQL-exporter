/*
* Samsung-cpc version 1.0
*
*  Copyright ⓒ 2023 kt corp. All rights reserved.
*
*  This is a proprietary software of kt corp, and you may not use this file except in
*  compliance with license agreement with kt corp. Any redistribution or use of this
*  software, with or without modification shall be strictly prohibited without prior written
*  approval of kt corp, and the copyright notice above does not evidence any actual or
*  intended publication of such software.
 */
package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"kt.com/p5g/cnf-exporter/samsung-cpc/cfg"
	"os"
)

var config = cfg.InitConfig()

// log 인스턴스
var log = newLogger(config.Logging.Encode)

func newLogger(encode string) *zap.Logger {
	cfg := os.Getenv(config.Logging.Level)

	var level zapcore.Level

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.LineEnding = zapcore.DefaultLineEnding
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.CallerKey = "caller"
	encoderCfg.MessageKey = "msg"

	switch cfg {
	case "INFO":
		level = zap.InfoLevel
	case "DEBUG":
		level = zap.DebugLevel
	case "FATAL":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          encode,
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}

	// AddCallerkskip을 해줘야 상위 method를 출력
	return zap.Must(config.Build(zap.AddCallerSkip(1)))
}

// fields사용법 : zap.String(Key,Value) 로 사용
// key Value는 String의 값으로 사용한다.
func LogDebug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

func LogInfo(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

func LogWarn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

func LogErr(msg string, err error) {
	log.Error(msg, zap.String("err", fmt.Sprint(err)))
}

func LogFatal(msg string, err error) {
	log.Fatal(msg, zap.String("err", fmt.Sprint(err)))
}
