package util

import (
	"fmt"
	"os"
	"regexp"

	"github.com/go-kit/kit/log"
	"google.golang.org/grpc/grpclog"
)

type (
	LogFn func(msg string)

	LogConfig struct {
		debug *debugLogger
		log   log.Logger
	}

	LogOpt func(*LogConfig)

	debugLogger struct {
		next    log.Logger
		debug   bool
		regexps []*regexp.Regexp
	}

	grpcLogToGoKitLog struct {
		l log.Logger
	}
)

func (g *grpcLogToGoKitLog) V(l int) bool {
	return false
}

func (g *grpcLogToGoKitLog) Info(args ...interface{}) {
	g.l.Log("info", fmt.Sprint(args...), "debug", true)
}

func (g *grpcLogToGoKitLog) Infoln(args ...interface{}) {
	g.l.Log("info", fmt.Sprint(args...), "debug", true)
}

func (g *grpcLogToGoKitLog) Infof(format string, args ...interface{}) {
	g.l.Log("info", fmt.Sprintf(format, args...), "debug", true)
}

func (g *grpcLogToGoKitLog) Warning(args ...interface{}) {
	g.l.Log("warn", fmt.Sprint(args...))
}

func (g *grpcLogToGoKitLog) Warningln(args ...interface{}) {
	g.l.Log("warn", fmt.Sprint(args...))
}

func (g *grpcLogToGoKitLog) Warningf(format string, args ...interface{}) {
	g.l.Log("warn", fmt.Sprintf(format, args...))
}

func (g *grpcLogToGoKitLog) Fatal(args ...interface{}) {
	g.l.Log("fatal", fmt.Sprint(args...))
	os.Exit(1)
}

func (g *grpcLogToGoKitLog) Fatalf(format string, args ...interface{}) {
	g.l.Log("fatal", fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (g *grpcLogToGoKitLog) Fatalln(args ...interface{}) {
	g.l.Log("fatal", fmt.Sprint(args...))
	os.Exit(1)
}

func (g *grpcLogToGoKitLog) Error(args ...interface{}) {
	g.l.Log("err", fmt.Sprint(args...))
}

func (g *grpcLogToGoKitLog) Errorf(format string, args ...interface{}) {
	g.l.Log("err", fmt.Sprintf(format, args...))
}

func (g *grpcLogToGoKitLog) Errorln(args ...interface{}) {
	g.l.Log("err", fmt.Sprint(args...))
}

// Should be modified only before first logger is created
// Also log messages with debug=true
var Debug bool = false

// Should be modified only before first logger is created
// Log debug messages with src matching one of provided regexps
var DebugFilters []string = nil

func GRPCLogToGoKitLog(l log.Logger) grpclog.LoggerV2 {
	return &grpcLogToGoKitLog{l: log.With(l, "src", "grpc", "debug", true)}
}

func SetupGRPCLogging(log log.Logger) {
	grpclog.SetLoggerV2(GRPCLogToGoKitLog(log))
}

type toStdErrWhenErrorSet struct {
	err log.Logger
	out log.Logger
}

func (l *toStdErrWhenErrorSet) Log(keyvals ...interface{}) error {
	for i := 0; i < len(keyvals); i += 2 {
		if k, ok := keyvals[i].(string); ok && k == "err" {
			return l.err.Log(keyvals...)
		}
	}

	return l.out.Log(keyvals...)
}

// Create new logger initialized based on Debug and DebugFilters vars
func Logger() log.Logger {
	stdErr := log.With(log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr)), "ts", log.DefaultTimestampUTC)
	stdOut := log.With(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), "ts", log.DefaultTimestampUTC)

	switched := &toStdErrWhenErrorSet{err: stdErr, out: stdOut}

	return makeDebugLogger(switched, Debug, DebugFilters...)
}

func makeDebugLogger(next log.Logger, debug bool, includeRegexps ...string) log.Logger {
	regexps := make([]*regexp.Regexp, len(includeRegexps))

	for i, r := range includeRegexps {
		regexps[i] = regexp.MustCompile(r)
	}

	return &debugLogger{
		debug:   debug,
		next:    next,
		regexps: regexps,
	}
}

func (d *debugLogger) Log(keyvals ...interface{}) error {
	if !d.debug {
		for i := 0; i < len(keyvals); i += 2 {
			if k, ok := keyvals[i].(string); ok && k == "debug" {
				return nil
			}
		}
	} else if len(d.regexps) > 0 {
		isDebugMsg := false
		isAcceptedComponent := false

		for i := 0; i < len(keyvals); i += 2 {
			if k, ok := keyvals[i].(string); ok {
				if k == "debug" {
					isDebugMsg = true
				} else if k == "src" && i+1 < len(keyvals) {
					src, ok := keyvals[i+1].(string)
					if ok {
						for _, r := range d.regexps {
							if r.MatchString(src) {
								isAcceptedComponent = true
								break
							}
						}
					}
				}
			}
		}

		if isDebugMsg && !isAcceptedComponent {
			return nil
		}
	}

	return d.next.Log(keyvals...)
}
