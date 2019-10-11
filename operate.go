package joker

import (
	"context"
)

// Debug uses fmt.Sprint to construct and log a message.
func Debug(args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Debug(args...)
	}
}

// Info uses fmt.Sprint to construct and log a message.
func Info(args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Info(args...)
	}
}

// Warn uses fmt.Sprint to construct and log a message.
func Warn(args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Warn(args...)
	}
}

// Error uses fmt.Sprint to construct and log a message.
func Error(args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Error(args...)
	}
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanic(args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.DPanic(args...)
	}
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func Panic(args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Panic(args...)
	}
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func Fatal(args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Fatal(args...)
	}
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debugf(template string, args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Debugf(template, args...)
	}
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Infof(template, args...)
	}
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(template string, args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Warnf(template, args...)
	}
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Errorf(template, args...)
	}
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanicf(template string, args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.DPanicf(template, args...)
	}
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func Panicf(template string, args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Panicf(template, args...)
	}
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func Fatalf(template string, args ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Fatalf(template, args...)
	}
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func Debugw(msg string, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Debugw(msg, keysAndValues...)
	}
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Infow(msg string, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Infow(msg, keysAndValues...)
	}
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Warnw(msg string, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Warnw(msg, keysAndValues...)
	}
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Errorw(msg string, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Errorw(msg, keysAndValues...)
	}
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func DPanicw(msg string, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.DPanicw(msg, keysAndValues...)
	}
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func Panicw(msg string, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Panicw(msg, keysAndValues...)
	}
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func Fatalw(msg string, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		defaultLogger.Fatalw(msg, keysAndValues...)
	}
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func Debugwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		keysAndValues = append(keysAndValues, requestIdKey, GetRequestId(ctx))
		defaultLogger.Debugw(msg, keysAndValues..., )
	}
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Infowc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		keysAndValues = append(keysAndValues, requestIdKey, GetRequestId(ctx))
		defaultLogger.Infow(msg, keysAndValues...)
	}
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Warnwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		keysAndValues = append(keysAndValues, requestIdKey, GetRequestId(ctx))
		defaultLogger.Warnw(msg, keysAndValues...)
	}
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Errorwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		keysAndValues = append(keysAndValues, requestIdKey, GetRequestId(ctx))
		defaultLogger.Errorw(msg, keysAndValues...)
	}
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func DPanicwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		keysAndValues = append(keysAndValues, requestIdKey, GetRequestId(ctx))
		defaultLogger.DPanicw(msg, keysAndValues...)
	}
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func Panicwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		keysAndValues = append(keysAndValues, requestIdKey, GetRequestId(ctx))
		defaultLogger.Panicw(msg, keysAndValues...)
	}
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func Fatalwc(msg string, ctx context.Context, keysAndValues ...interface{}) {
	if defaultLoggerStatus {
		keysAndValues = append(keysAndValues, requestIdKey, GetRequestId(ctx))
		defaultLogger.Fatalw(msg, keysAndValues...)
	}
}

func Sync() {
	defaultLogger.With()
	if defaultLoggerStatus {
		defaultLogger.Sync()
	}
}
