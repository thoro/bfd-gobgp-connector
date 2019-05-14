//
// Simple logging facility with daily rotating logfiles.
//
// Each line of the log has a line prefix, composed as following:
//   [<level> <date>-<time> <pid> <file>:<line>] <message>
//
// So a single log entry might look like this:
//   [INFO 0509-17:16:17 18191 proxy.go:28] Listening on :5432
//

package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	LevelInfo LogLevel = iota
	LevelWarn
	LevelError
)

var levelName = []string{
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

type Logger struct {
	filePath     string
	toStdout     bool
	nextRotation time.Time
	file         *os.File
	mutex        sync.Mutex
}

var logger Logger

func init() {
	logger.filePath = "proxy.log"
}

func SetLogfileName(path string) {
	logger.filePath = path
}

func SetLogToStdout(value bool) {
	logger.toStdout = value
}

func (l *Logger) output(level LogLevel, format string, args ...interface{}) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.file == nil {
		if err := l.openFile(); err != nil {
			fmt.Println(err)
			return 0, err
		}
	}

	if time.Now().After(l.nextRotation) {
		if err := l.rotateFiles(); err != nil {
			fmt.Println(err)
			return 0, err
		}
	}

	var sb strings.Builder
	_, err := sb.WriteString(linePrefix(level))
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	if len(format) > 0 {
		fmt.Fprintf(&sb, format, args...)
	} else {
		fmt.Fprint(&sb, args...)
	}

	written, err := l.writeOut(sb.String())
	if err != nil {
		fmt.Println(err)
	}
	return written, err
}

func (l *Logger) writeOut(message string) (int, error) {
	needNewline := message[len(message)-1] != '\n'

	if l.toStdout {
		os.Stdout.WriteString(message)
		if needNewline {
			os.Stdout.WriteString("\n")
		}
	}

	written, err := l.file.WriteString(message)
	if needNewline {
		l.file.WriteString("\n")
	}
	return written, err
}

func (l *Logger) openFile() (err error) {
	path := l.filename()
	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}

	l.file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}

	l.nextRotation = nextRotationTime()
	return
}

func (l *Logger) rotateFiles() error {
	l.file.Close()
	l.file = nil

	return l.openFile()
}

// Returns a relative filepath to where the logfile should be written to,
// e.g. log/<name-of-executable>-<year>-<month>-<day>.log
func (l *Logger) filename() string {
	now := time.Now()

	ext := filepath.Ext(l.filePath)
	base := l.filePath[0 : len(l.filePath)-len(ext)]

	return fmt.Sprintf(
		"%s-%d-%d-%d%s",
		base, now.Year(), now.Month(), now.Day(), ext,
	)
}

func linePrefix(loglevel LogLevel) string {
	now := time.Now().Format("0102-15:04:05")
	pid := os.Getpid()
	level := levelName[loglevel]
	file, line := callerFileInfo()

	return fmt.Sprintf("[%s %s %d %s:%d] ", level, now, pid, file, line)
}

// Returns filename and linenumber of the caller of any of {Info,Warn,Error}[f]
func callerFileInfo() (string, int) {
	// At this very point, the calling depth is four.
	//   #0 callerFileInfo()
	//   #1 linePrefix()
	//   #2 Logger.output()
	//   #3 {Info,Warn,Error}[f]
	const CallingDepth = 4

	_, file, line, ok := runtime.Caller(CallingDepth)

	if !ok {
		file = "<unknown>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}

	return file, line
}

// Returns the point in time when the log should be rotated next time.
// Currently: start of next day
func nextRotationTime() time.Time {
	t := time.Now().AddDate(0, 0, 1)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func Info(args ...interface{}) {
	logger.output(LevelInfo, "", args...)
}

func Infof(format string, args ...interface{}) {
	logger.output(LevelInfo, format, args...)
}

func Warn(args ...interface{}) {
	logger.output(LevelWarn, "", args...)
}

func Warnf(format string, args ...interface{}) {
	logger.output(LevelWarn, format, args...)
}

func Error(args ...interface{}) {
	logger.output(LevelError, "", args...)
}

func Errorf(format string, args ...interface{}) {
	logger.output(LevelError, format, args...)
}
