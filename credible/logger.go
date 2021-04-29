package credible

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type (
	Logger interface {
		Debug(interface{})
		Error(interface{})
		Debugf(string, ...interface{})
		Errorf(string, ...interface{})
		//RedirectLevel out the log to the specified target
		RedirectLevel(string, io.ReadWriteCloser)
	}

	fmtJson struct {
		Level string      `json:"level"`
		Time  string      `json:"time"`
		File  string      `json:"file"`
		Func  string      `json:"func"`
		Line  int         `json:"line"`
		Msg   interface{} `json:"data"`
	}

	logger struct {
		levelOut   map[string]io.ReadWriteCloser
		levelMutex map[string]*sync.RWMutex
		timeFormat string
	}
)

const (
	LEVEL_DEBUG = "debug"
	LEVEL_ERROR = "error"
)

var (
	_ Logger = &logger{}

	globLog = newLogger()
)

func newLogger() Logger {
	log := &logger{
		levelOut:   make(map[string]io.ReadWriteCloser),
		levelMutex: make(map[string]*sync.RWMutex),
		timeFormat: "2006/01/02 15:04:05",
	}

	log.levelOut[LEVEL_DEBUG] = os.Stdout
	log.levelMutex[LEVEL_DEBUG] = &sync.RWMutex{}
	log.levelOut[LEVEL_ERROR] = os.Stderr
	log.levelMutex[LEVEL_ERROR] = &sync.RWMutex{}

	return log
}

func NewFile(fileName string) io.ReadWriteCloser {
	pwd, err := os.Getwd()
	if nil != err {
		pwd = "."
	}
	dir := filepath.Join(pwd, "log")
	if _, err = os.Stat(dir); nil != err {
		err = os.MkdirAll(dir, os.ModePerm)
		if nil != err {
			return os.Stderr
		}
	}
	fullPath := filepath.Join(dir, fileName)
	if _, err := os.Stat(fullPath); nil == err {
		os.Rename(fullPath, fullPath+".old")
	}

	f, err := os.Create(fullPath)
	if nil != err {
		return os.Stderr
	}

	return f
}

func (my *logger) Debug(data interface{}) {
	pc, file, line, _ := runtime.Caller(2)
	tnow := time.Now().Format(my.timeFormat)

	para := &fmtJson{
		Level: LEVEL_DEBUG,
		Time:  tnow,
		File:  filepath.Base(file),
		Func:  runtime.FuncForPC(pc).Name(),
		Line:  line,
		Msg:   data,
	}

	bs, err := json.Marshal(para)
	my.levelMutex[LEVEL_DEBUG].Lock()
	defer my.levelMutex[LEVEL_DEBUG].Unlock()
	if nil != err {
		io.WriteString(os.Stderr, err.Error()+"\n")
	} else {
		io.WriteString(my.levelOut[LEVEL_DEBUG], string(bs)+"\n")
	}

}

func (my *logger) Error(data interface{}) {
	pc, file, line, _ := runtime.Caller(2)
	tnow := time.Now().Format(my.timeFormat)

	para := &fmtJson{
		Level: LEVEL_ERROR,
		Time:  tnow,
		File:  filepath.Base(file),
		Func:  runtime.FuncForPC(pc).Name(),
		Line:  line,
		Msg:   data,
	}

	bs, err := json.Marshal(para)
	my.levelMutex[LEVEL_ERROR].Lock()
	defer my.levelMutex[LEVEL_ERROR].Unlock()
	if nil != err {
		io.WriteString(os.Stderr, err.Error())
	} else {
		io.WriteString(my.levelOut[LEVEL_ERROR], string(bs)+"\n")
	}
}

func (my *logger) Debugf(fmtstr string, v ...interface{}) {
	pc, file, line, _ := runtime.Caller(2)
	tnow := time.Now().Format(my.timeFormat)

	para := &fmtJson{
		Level: LEVEL_DEBUG,
		Time:  tnow,
		File:  filepath.Base(file),
		Func:  runtime.FuncForPC(pc).Name(),
		Line:  line,
		Msg:   fmt.Sprintf(fmtstr, v...),
	}

	bs, err := json.Marshal(para)
	my.levelMutex[LEVEL_DEBUG].Lock()
	defer my.levelMutex[LEVEL_DEBUG].Unlock()
	if nil != err {
		io.WriteString(os.Stderr, err.Error()+"\n")
	} else {
		io.WriteString(my.levelOut[LEVEL_DEBUG], string(bs)+"\n")
	}
}

func (my *logger) Errorf(fmtstr string, v ...interface{}) {
	pc, file, line, _ := runtime.Caller(2)
	tnow := time.Now().Format(my.timeFormat)

	para := &fmtJson{
		Level: LEVEL_ERROR,
		Time:  tnow,
		File:  filepath.Base(file),
		Func:  runtime.FuncForPC(pc).Name(),
		Line:  line,
		Msg:   fmt.Sprintf(fmtstr, v...),
	}

	bs, err := json.Marshal(para)
	my.levelMutex[LEVEL_ERROR].Lock()
	defer my.levelMutex[LEVEL_ERROR].Unlock()
	if nil != err {
		io.WriteString(os.Stderr, err.Error())
	} else {
		io.WriteString(my.levelOut[LEVEL_ERROR], string(bs)+"\n")
	}
}

func (my *logger) RedirectLevel(level string, out io.ReadWriteCloser) {
	my.levelMutex[level].Lock()
	defer my.levelMutex[level].Unlock()
	my.levelOut[level] = out
}

func Debug(d interface{}) {
	globLog.Debug(d)
}

func Debugf(fmtStr string, v ...interface{}) {
	globLog.Debugf(fmtStr, v...)
}

func Error(d interface{}) {
	globLog.Error(d)
}

func Errorf(fmtStr string, v ...interface{}) {
	globLog.Errorf(fmtStr, v...)
}

func RedirectLevel(level string, out io.ReadWriteCloser) {
	globLog.RedirectLevel(level, out)
}
