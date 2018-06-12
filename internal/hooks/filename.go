package hooks

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"sync"
)

type Hook struct {
	sync.RWMutex
	Field  string
	levels []logrus.Level
}

func (hook *Hook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *Hook) Fire(entry *logrus.Entry) error {
	hook.Lock()
	defer hook.Unlock()

	entry.Data[hook.Field] = findCaller()
	return nil
}

func NewHook(levels ...logrus.Level) *Hook {
	hook := Hook{
		Field:  "source",
		levels: levels,
	}
	if len(hook.levels) == 0 {
		hook.levels = logrus.AllLevels
	}

	return &hook
}

func findCaller() string {
	file := ""
	line := 0
	skip := 5
	for i := 0; i < 10; i++ {
		file, line = getCaller(skip + i)
		if !strings.Contains(file, "logrus") {
			break
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func getCaller(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0
	}

	//n := 0
	//for i := len(file) - 1; i > 0; i-- {
	//	if file[i] == '/' {
	//		n += 1
	//		if n >= 2 {
	//			file = file[i+1:]
	//			break
	//		}
	//	}
	//}

	return file, line
}
