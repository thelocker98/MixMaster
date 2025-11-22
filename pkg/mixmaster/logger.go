package mixmaster

import (
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(buildType string, cfg *Config) (*zap.SugaredLogger, error) {
	var loggerConfig zap.Config

	if buildType == "release" {

		loggerConfig = zap.NewProductionConfig()

		loggerConfig.OutputPaths = []string{filepath.Join(cfg.App.LogFileLocation, cfg.App.LogFileName)}
		loggerConfig.Encoding = "console"

	} else {
		loggerConfig = zap.NewDevelopmentConfig()

		loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	loggerConfig.EncoderConfig.EncodeCaller = nil
	loggerConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}

	loggerConfig.EncoderConfig.EncodeName = func(s string, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("%-27s", s))
	}

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("create zap logger: %w", err)
	}

	sugar := logger.Sugar()

	return sugar, nil
}
