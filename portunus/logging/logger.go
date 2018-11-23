// Copyright © 2018 Carl P. Corliss <carl@corliss.name>
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package logging

import (
	"io"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

var oneLogInstance sync.Once
var logInstance *logrus.Logger

type Fields = logrus.Fields

func newLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		DisableSorting:         true,
	})

	logLevel := config.GetString("logging.level")
	if level, err := logrus.ParseLevel(logLevel); err != nil {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(level)
	}

	return logger
}

func Writer() *io.PipeWriter {
	return GetLogger().Writer()
}

func GetLogger() *logrus.Logger {
	oneLogInstance.Do(func() {
		logInstance = newLogger()
	})
	return logInstance
}

func extractFields(args []interface{}) (logrus.Fields, []interface{}) {
	if fieldMap, ok := args[len(args)-1].(logrus.Fields); ok {
		args = args[:len(args)-1]
		return fieldMap, args
	}
	return nil, args
}

func IsDebugEnabled() bool { return GetLogger().IsLevelEnabled(logrus.DebugLevel) }
func IsErrorEnabled() bool { return GetLogger().IsLevelEnabled(logrus.ErrorLevel) }
func IsFatalEnabled() bool { return GetLogger().IsLevelEnabled(logrus.FatalLevel) }
func IsInfoEnabled() bool  { return GetLogger().IsLevelEnabled(logrus.InfoLevel) }
func IsPanicEnabled() bool { return GetLogger().IsLevelEnabled(logrus.PanicLevel) }
func IsTraceEnabled() bool { return GetLogger().IsLevelEnabled(logrus.TraceLevel) }
func IsWarnEnabled() bool  { return GetLogger().IsLevelEnabled(logrus.WarnLevel) }

func Tracef(format string, args ...interface{}) {
	logrus.WithFields(retrieveCallInfo()).Tracef(format, args...)
}
func Debugf(format string, args ...interface{}) {
	logrus.WithFields(retrieveCallInfo()).Debugf(format, args...)
}
func Infof(format string, args ...interface{})    { logrus.Infof(format, args...) }
func Printf(format string, args ...interface{})   { logrus.Printf(format, args...) }
func Warnf(format string, args ...interface{})    { logrus.Warnf(format, args...) }
func Warningf(format string, args ...interface{}) { logrus.Warningf(format, args...) }
func Errorf(format string, args ...interface{}) {
	logrus.WithFields(retrieveCallInfo()).Errorf(format, args...)
}
func Fatalf(format string, args ...interface{}) {
	logrus.WithFields(retrieveCallInfo()).Fatalf(format, args...)
}
func Panicf(format string, args ...interface{}) {
	logrus.WithFields(retrieveCallInfo()).Panicf(format, args...)
}
func Trace(args ...interface{})     { logrus.WithFields(retrieveCallInfo()).Trace(args...) }
func Debug(args ...interface{})     { logrus.WithFields(retrieveCallInfo()).Debug(args...) }
func Info(args ...interface{})      { logrus.Info(args...) }
func Print(args ...interface{})     { logrus.Print(args...) }
func Warn(args ...interface{})      { logrus.Warn(args...) }
func Warning(args ...interface{})   { logrus.Warning(args...) }
func Error(args ...interface{})     { logrus.WithFields(retrieveCallInfo()).Error(args...) }
func Fatal(args ...interface{})     { logrus.WithFields(retrieveCallInfo()).Fatal(args...) }
func Panic(args ...interface{})     { logrus.WithFields(retrieveCallInfo()).Panic(args...) }
func Traceln(args ...interface{})   { logrus.WithFields(retrieveCallInfo()).Traceln(args...) }
func Debugln(args ...interface{})   { logrus.WithFields(retrieveCallInfo()).Debugln(args...) }
func Infoln(args ...interface{})    { logrus.Infoln(args...) }
func Println(args ...interface{})   { logrus.Println(args...) }
func Warnln(args ...interface{})    { logrus.Warnln(args...) }
func Warningln(args ...interface{}) { logrus.Warningln(args...) }
func Errorln(args ...interface{})   { logrus.WithFields(retrieveCallInfo()).Errorln(args...) }
func Fatalln(args ...interface{})   { logrus.WithFields(retrieveCallInfo()).Fatalln(args...) }
func Panicln(args ...interface{})   { logrus.WithFields(retrieveCallInfo()).Panicln(args...) }

func TracefWithFields(format string, args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.TraceLevel) {
		entry := logger.WithFields(retrieveCallInfo())
		if fields, args := extractFields(args); fields != nil {
			entry.WithFields(fields).Tracef(format, args...)
		} else {
			entry.Tracef(format, args...)
		}
	}
}

func DebugfWithFields(format string, args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.DebugLevel) {
		entry := logger.WithFields(retrieveCallInfo())
		if fields, args := extractFields(args); fields != nil {
			entry.WithFields(fields).Debugf(format, args...)
		} else {
			entry.Debugf(format, args...)
		}
	}
}

func InfofWithFields(format string, args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.InfoLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Infof(format, args...)
		} else {
			logger.Infof(format, args...)
		}
	}
}

func PrintfWithFields(format string, args ...interface{}) {
	logger := GetLogger()
	if fields, args := extractFields(args); fields != nil {
		logger.WithFields(fields).Printf(format, args...)
	} else {
		logger.Printf(format, args...)
	}
}

func WarnfWithFields(format string, args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.WarnLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Warnf(format, args...)
		} else {
			logger.Warnf(format, args...)
		}
	}
}

func WarningfWithFields(format string, args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.WarnLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Warnf(format, args...)
		} else {
			logger.Warnf(format, args...)
		}
	}
}

func ErrorfWithFields(format string, args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.ErrorLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Errorf(format, args...)
		} else {
			logger.Errorf(format, args...)
		}
	}
}

func FatalfWithFields(format string, args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.FatalLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Fatalf(format, args...)
		} else {
			logger.Fatalf(format, args...)
		}
	}
	logger.Exit(1)
}

func PanicfWithFields(format string, args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.PanicLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Panicf(format, args...)
		} else {
			logger.Panicf(format, args...)
		}
	}
}

func TraceWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.TraceLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Trace(args...)
		} else {
			logger.Trace(args...)
		}
	}
}

func DebugWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.DebugLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Debug(args...)
		} else {
			logger.Debug(args...)
		}
	}
}

func InfoWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.InfoLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Info(args...)
		} else {
			logger.Info(args...)
		}
	}
}

func PrintWithFields(args ...interface{}) {
	logger := GetLogger()
	if fields, args := extractFields(args); fields != nil {
		logger.WithFields(fields).Info(args...)
	} else {
		logger.Info(args...)
	}
}

func WarnWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.WarnLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Warn(args...)
		} else {
			logger.Warn(args...)
		}
	}
}

func WarningWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.WarnLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Warn(args...)
		} else {
			logger.Warn(args...)
		}
	}
}

func ErrorWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.ErrorLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Error(args...)
		} else {
			logger.Error(args...)
		}
	}
}

func FatalWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.FatalLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Fatal(args...)
		} else {
			logger.Fatal(args...)
		}
	}
	logger.Exit(1)
}

func PanicWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.PanicLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Panic(args...)
		} else {
			logger.Panic(args...)
		}
	}
}

func TracelnWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.TraceLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Traceln(args...)
		} else {
			logger.Traceln(args...)
		}
	}
}

func DebuglnWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.DebugLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Debugln(args...)
		} else {
			logger.Debugln(args...)
		}
	}
}

func InfolnWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.InfoLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Infoln(args...)
		} else {
			logger.Infoln(args...)
		}
	}
}

func PrintlnWithFields(args ...interface{}) {
	logger := GetLogger()
	if fields, args := extractFields(args); fields != nil {
		logger.WithFields(fields).Println(args...)
	} else {
		logger.Println(args...)
	}
}

func WarnlnWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.WarnLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Warnln(args...)
		} else {
			logger.Warnln(args...)
		}
	}
}

func WarninglnWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.WarnLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Warnln(args...)
		} else {
			logger.Warnln(args...)
		}
	}
}

func ErrorlnWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.ErrorLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Errorln(args...)
		} else {
			logger.Errorln(args...)
		}
	}
}

func FatallnWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.FatalLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Fatalln(args...)
		} else {
			logger.Fatalln(args...)
		}
	}
	logger.Exit(1)
}

func PaniclnWithFields(args ...interface{}) {
	logger := GetLogger()
	if logger.IsLevelEnabled(logrus.PanicLevel) {
		if fields, args := extractFields(args); fields != nil {
			logger.WithFields(fields).Panicln(args...)
		} else {
			logger.Panicln(args...)
		}
	}
}

// -------------------------------------------------------------------------- //
// Below functionality (though modified slightly) is borrowed from the
// BSD-3 licensed Tideland Go Library, which can be found at:
// @see https://github.com/tideland/golib/blob/master/logger/logger.go#L516
// -------------------------------------------------------------------------- //

// callInfo bundles the info about the call environment
// when a logging statement occurred.

// Retrieves caller details, including package, file, and line numbers
func retrieveCallInfo() logrus.Fields {
	pc, file, line, _ := runtime.Caller(2)
	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	packageName := ""
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		packageName = strings.Join(parts[0:pl-2], ".")
	} else {
		packageName = strings.Join(parts[0:pl-1], ".")
	}

	return logrus.Fields{
		"packageName": packageName,
		"fileName":    fileName,
		"funcName":    funcName,
		"line":        line,
	}
}
