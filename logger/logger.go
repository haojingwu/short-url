package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

// InitLogger 初始化zap日志
func InitLogger() {
	//1配置日志输出格式
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime:    zapcore.ISO8601TimeEncoder,

		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, //显示文件名:行号
	}

	//2配置日志输出目标(控制台+文件)
	//控制台输出(开发环境方便查看)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleWriter := zapcore.AddSync(os.Stdout)

	//文件输出(生产环境持久化)
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log", //日志文件路径
		MaxSize:    10,             //每个日志文件最大10MB
		MaxBackups: 30,             //保留30个旧文件
		MaxAge:     7,              //保留7天
		Compress:   true,           //压缩旧文件
	})

	//3创建Core(同时输出到控制台和文件)
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleWriter, zapcore.DebugLevel),
		zapcore.NewCore(consoleEncoder, fileWriter, zapcore.InfoLevel),
	)

	//4创建Logger
	Log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	Log.Info("✅ Logger 初始化成功")
}

// Sync刷新日志缓冲区
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// 便捷方法,方便其他包调用
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}
