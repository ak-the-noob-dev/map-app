package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// getLogFilePath generates a log file path with the current date
func getLogFilePath(logType string) string {
	currentDate := time.Now().Format("2006-01-02") // YYYY-MM-DD format
	return "./logs/" + logType + "-" + currentDate + ".log"
}
 
// NewLogger creates a single logger that handles both general and error logs
func NewLogger() *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
 
	// Log rotation settings for general logs
	logFile := &lumberjack.Logger{
		Filename:   getLogFilePath("application"), // Daily rotated log
		MaxSize:    10,                            // Max file size in MB before rotation
		MaxAge:     7,                             // Keep logs for 7 days
		MaxBackups: 7,                             // Keep max 7 backup files
		Compress:   true,                          // Compress old log files
		LocalTime:  true,
	}
 
	// Log rotation settings for error logs
	errorLogFile := &lumberjack.Logger{
		Filename:   getLogFilePath("error"), // Daily rotated error log
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 7,
		Compress:   true,
		LocalTime:  true,
	}
 
	// Create core for different log levels
	consoleWriter := zapcore.AddSync(os.Stderr)
	fileWriter := zapcore.AddSync(logFile)
	errorFileWriter := zapcore.AddSync(errorLogFile)
 
	// Define log level enablers
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel // Log only errors and above
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel // Log only info and warnings
	})
 
	// Create cores for general and error logs
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), fileWriter, lowPriority),          // Info/Warn -> application.log
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), errorFileWriter, highPriority),    // Errors -> error.log
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), consoleWriter, zap.DebugLevel), // Print everything to stderr
	)
 
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
 