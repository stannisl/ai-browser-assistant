package logger

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	infoColor       = color.New(color.FgCyan)
	debugColor      = color.New(color.FgGreen)
	errorColor      = color.New(color.FgRed)
	warnColor       = color.New(color.FgYellow)
	successColor    = color.New(color.FgGreen)
	errorMsgColor   = color.New(color.FgWhite, color.Bold)
	successMsgColor = color.New(color.FgGreen, color.Bold)
	stepColor       = color.New(color.FgMagenta)
	toolColor       = color.New(color.FgBlue)
	thinkColor      = color.New(color.FgYellow)
)

type Logger struct {
	sugared *zap.SugaredLogger
}

func New(debug bool) (*Logger, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	if debug {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.Development = true
	}

	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.NameKey = "logger"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.FunctionKey = zapcore.OmitKey
	config.EncoderConfig.MessageKey = "msg"
	config.EncoderConfig.StacktraceKey = "stacktrace"
	config.EncoderConfig.LineEnding = zapcore.DefaultLineEnding
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	encoderConfig := config.EncoderConfig
	encoderConfig.TimeKey = "time"
	encoderConfig.LevelKey = "level"
	encoderConfig.NameKey = "logger"
	encoderConfig.CallerKey = "caller"
	encoderConfig.FunctionKey = zapcore.OmitKey
	encoderConfig.MessageKey = "msg"
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.LineEnding = zapcore.DefaultLineEnding
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	var logger *zap.Logger
	var err error
	if debug {
		logger, err = config.Build(zap.AddCaller(), zap.Development())
		if err != nil {
			return nil, fmt.Errorf("failed to create logger in debug mode: %w", err)
		}
	} else {
		logger, err = config.Build(zap.AddCaller())
		if err != nil {
			return nil, fmt.Errorf("failed to create logger: %w", err)
		}
	}

	return &Logger{sugared: logger.Sugar()}, nil
}

func (l *Logger) Close() {
	l.sugared.Sync()
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.sugared.Infow(msg, keysAndValues...)
}

func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.sugared.Debugw(msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.sugared.Errorw(msg, keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.sugared.Warnw(msg, keysAndValues...)
}

func (l *Logger) Extract(url string, count int) {
	infoColor.Printf("🔍 [EXTRACT] %s (%d elements)\n", truncate(url, 100), count)
	l.sugared.Info("Extract page content", "url", url, "count", count)
}

func (l *Logger) Navigate(url string) {
	infoColor.Printf("🌐 [NAVIGATE] %s\n", truncate(url, 100))
	l.sugared.Info("Navigate to URL", "url", url)
}

func (l *Logger) Click(id int, text string) {
	infoColor.Printf("🖱️  [CLICK] [%d] %s\n", id, truncate(text, 50))
	l.sugared.Info("Click element", "id", id, "text", text)
}

func (l *Logger) Type(id int, text string) {
	infoColor.Printf("⌨️  [TYPE] [%d] \"%s\"\n", id, truncate(text, 50))
	l.sugared.Info("Type into element", "id", id, "text", text)
}

func (l *Logger) Confirm(desc string) {
	errorMsgColor.Printf("🔒 [CONFIRM] %s\n", truncate(desc, 100))
	l.sugared.Warn("Confirmation required", "description", desc)
}

func (l *Logger) Done(msg string, success bool) {
	if success {
		successMsgColor.Printf("✅ [DONE] %s\n", truncate(msg, 100))
		l.sugared.Info("Operation completed successfully", "message", msg)
	} else {
		errorMsgColor.Printf("❌ [FAILED] %s\n", truncate(msg, 100))
		l.sugared.Error("Operation failed", "message", msg)
	}
}

func (l *Logger) Thinking() {
	thinkColor.Print("🤔 [THINKING]...\n")
	l.sugared.Debug("AI is thinking")
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func init() {
	if !isTerminal() {
		disableColors()
	}
}

func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func disableColors() {
	color.NoColor = true
}

func (l *Logger) Tool(name string) {
	toolColor.Printf("🔧 [TOOL] %s\n", name)
	l.sugared.Infow("Executing tool", "tool", name)
}

func (l *Logger) Step(current, max int) {
	stepColor.Printf("📍 [STEP %d/%d]\n", current, max)
	l.sugared.Infow("Step progress", "current", current, "max", max)
}

func (l *Logger) Scroll(direction string) {
	infoColor.Printf("📜 [SCROLL] %s\n", direction)
	l.sugared.Infow("Scroll page", "direction", direction)
}

func (l *Logger) Ask(question string) {
	infoColor.Printf("💬 [ASK] %s\n", truncate(question, 100))
	l.sugared.Infow("User question", "question", question)
}
