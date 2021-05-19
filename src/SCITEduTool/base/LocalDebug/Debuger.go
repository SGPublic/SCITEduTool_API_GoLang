package LocalDebug

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

//func IsDebug() bool {
//	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
//	if err != nil {
//		return false
//	}
//	return strings.Contains(path, "Documents")
//}

func CheckLogDir() (bool, string) {
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
