package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L *zap.Logger

func Init(serviceName string) {
	// 从环境变量读取
	env := os.Getenv("APP_ENV")
	levelStr := os.Getenv("LOG_LEVEL")

	// 环境判断
	isProd := env == "production"

	// 日志级别判断
	var atomLevel zap.AtomicLevel
	switch levelStr {
	case "debug":
		atomLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	default:
		atomLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if isProd {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), atomLevel)

	L = zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel)).With(zap.String("service", serviceName))
}

func FromContext(ctx context.Context) *zap.Logger {
	if L == nil {
		return zap.NewNop()
	}

	traceID, _ := ctx.Value("trace_id").(string)
	tenantID, _ := ctx.Value("tenant_id").(string)
	userID, _ := ctx.Value("user_id").(string)

	var fields []zap.Field
	if traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if tenantID != "" {
		fields = append(fields, zap.String("tenant_id", tenantID))
	}

	if userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}
	return L.With(fields...)
}
