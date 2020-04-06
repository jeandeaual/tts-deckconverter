package log

// A global variable so that log functions can be directly accessed
var log Logger

// Logger is a logger abstraction
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
}

// SetLogger sets the logger instance used by the package.
func SetLogger(logger Logger) {
	log = logger
}

// Debug uses fmt.Sprint to construct and log a message.
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Info uses fmt.Sprint to construct and log a message.
func Info(args ...interface{}) {
	log.Info(args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Error uses fmt.Sprint to construct and log a message.
func Error(args ...interface{}) {
	log.Error(args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func Panic(args ...interface{}) {
	log.Panic(args...)
}

// Debugf uses fmt.Sprintf to construct and log a message.
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Infof uses fmt.Sprintf to construct and log a message.
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warnf uses fmt.Sprintf to construct and log a message.
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Errorf uses fmt.Sprintf to construct and log a message.
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Fatalf uses fmt.Sprint to construct and log a message, then calls os.Exit.
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// Panicf uses fmt.Sprint to construct and log a message, then panics.
func Panicf(format string, args ...interface{}) {
	log.Panicf(format, args...)
}

// Debugw logs a message with some additional context.
func Debugw(msg string, keysAndValues ...interface{}) {
	log.Debugw(msg, keysAndValues...)
}

// Infow logs a message with some additional context.
func Infow(msg string, keysAndValues ...interface{}) {
	log.Infow(msg, keysAndValues...)
}

// Warnw logs a message with some additional context.
func Warnw(msg string, keysAndValues ...interface{}) {
	log.Warnw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context.
func Errorw(msg string, keysAndValues ...interface{}) {
	log.Errorw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics.
func Panicw(msg string, keysAndValues ...interface{}) {
	log.Panicw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit.
func Fatalw(msg string, keysAndValues ...interface{}) {
	log.Fatalw(msg, keysAndValues...)
}
