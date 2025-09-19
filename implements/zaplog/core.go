package zaplog

import (
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newCore(writer io.Writer) zapcore.Core {
	return zapcore.NewCore(getCoreEncoder(), zapcore.AddSync(writer), zapcore.DebugLevel)
}

func getCoreEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderConfig.TimeKey = "time"
	return zapcore.NewJSONEncoder(encoderConfig)
}
