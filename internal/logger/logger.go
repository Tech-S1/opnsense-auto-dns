package logger

import (
	"os"
	"runtime"
	"strings"

	"github.com/charmbracelet/log"
)

var globalLogger *log.Logger

func Init(level log.Level) {
	globalLogger = log.NewWithOptions(os.Stdout, log.Options{
		Level: level,
	})
}

func SetLevel(level log.Level) {
	if globalLogger != nil {
		globalLogger = log.NewWithOptions(os.Stdout, log.Options{
			Level: level,
		})
	}
}

func Get() *log.Logger {
	if globalLogger == nil {
		Init(log.InfoLevel)
	}
	return globalLogger
}

func getCallerInfo() (string, string, string, int) {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown", "unknown", "unknown", 0
	}

	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	packageName := "unknown"
	if fn != nil {
		fullName := fn.Name()
		parts := strings.Split(fullName, ".")
		if len(parts) >= 2 {
			funcName = parts[len(parts)-1]
			packagePath := strings.Join(parts[:len(parts)-1], ".")
			pkgParts := strings.Split(packagePath, "/")
			if len(pkgParts) > 0 {
				packageName = pkgParts[len(pkgParts)-1]
			}
		} else if len(parts) == 1 {
			funcName = parts[0]
		}
	}

	parts := strings.Split(file, "/")
	fileName := "unknown"
	if len(parts) > 0 {
		fileName = parts[len(parts)-1]
	}

	return funcName, packageName, fileName, line
}

func Debug(msg string, args ...any) {
	funcName, packageName, fileName, line := getCallerInfo()
	callerArgs := append([]any{"func", funcName, "pkg", packageName, "file", fileName, "line", line}, args...)
	Get().Debug(msg, callerArgs...)
}

func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	funcName, packageName, fileName, line := getCallerInfo()
	callerArgs := append([]any{"func", funcName, "pkg", packageName, "file", fileName, "line", line}, args...)
	Get().Error(msg, callerArgs...)
}

func Fatal(msg string, args ...any) {
	funcName, packageName, fileName, line := getCallerInfo()
	callerArgs := append([]any{"func", funcName, "pkg", packageName, "file", fileName, "line", line}, args...)
	Get().Fatal(msg, callerArgs...)
}
