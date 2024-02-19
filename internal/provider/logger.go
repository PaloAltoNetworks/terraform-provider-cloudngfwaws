package provider

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type Level = zapcore.Level

const (
	InfoLevel  Level = zap.InfoLevel  // 0, default level
	WarnLevel  Level = zap.WarnLevel  // 1
	ErrorLevel Level = zap.ErrorLevel // 2
	DebugLevel Level = zap.DebugLevel // -1
)

var Logger *zap.SugaredLogger

func InitLogger(level Level) {
	writerSyncer := getLogWriter()
	stdoutSyncer := zapcore.Lock(os.Stdout)

	encoder := getEncoder()
	core := zapcore.NewTee(
		zapcore.NewCore(
			encoder,
			writerSyncer,
			level,
		),
		zapcore.NewCore(
			encoder,
			stdoutSyncer,
			level,
		),
	)
	logger := zap.New(core, zap.AddCaller())
	Logger = logger.Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "/var/log/terraform.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}
