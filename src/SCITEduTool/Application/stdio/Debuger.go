package stdio

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var debug = false

type localDebug interface {
	IsDebug() bool
	SetupDebugConfig(isDebug bool)
	CheckLogDir() (bool, string)
}

type localDebugImpl struct{}

var LocalDebug localDebug = localDebugImpl{}

func (localDebugImpl localDebugImpl) IsDebug() bool {
	return debug
}

func (localDebugImpl localDebugImpl) SetupDebugConfig(isDebug bool) {
	debug = isDebug
}

func (localDebugImpl localDebugImpl) CheckLogDir() (bool, string) {
	path, err := getLogDir()
	_, err = os.Stat(path + "user")
	if err == nil {
		return true, path
	}
	if !os.IsNotExist(err) {
		return false, ""
	}
	err = os.MkdirAll(path+"user", 0644)
	return err == nil, path
}

func getLogDir() (string, error) {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	now := time.Now()
	dateString := strconv.Itoa(now.Year()) + "_" + now.Month().String() + "_" + strconv.Itoa(now.Day())
	path += "/log/" + dateString + "/"
	return path, nil
}
