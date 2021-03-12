package StdOutUnit

import (
	"SCITEduTool/base/LocalDebug"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type MessagedError struct {
	Code       int
	HasInfo    bool
	OutMessage func(w http.ResponseWriter)
}

type StringMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MyLog interface {
	String(str string)
	Object(obj interface{})
}

func outInTerminal(username string, prefix string, str string, flag int, calldepth int) {
	if username != "" {
		str = "[" + username + "] " + str
	}
	doOut(os.Stdout, prefix, str+"\x1b[0m", flag, calldepth)
}

func outInFile(username string, prefix string, str string, flag int, calldepth int) {
	if LocalDebug.IsDebug() {
		return
	}
	exist, path := LocalDebug.CheckLogDir()
	if !exist {
		return
	}
	if username != "" {
		path += "user/" + username + ".log"
	} else {
		path += "dereliction.log"
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	doOut(io.Writer(file), prefix, str, flag, calldepth)
}

func doOut(writer io.Writer, prefix string, str string, flag int, calldepth int) {
	MyLog := log.New(writer, prefix, flag)
	_ = MyLog.Output(calldepth, str)
}

func Verbose(username string, str string) {
	if !LocalDebug.IsDebug() {
		return
	}
	outInTerminal(username, "\x1B[1;37;1m[Verbose] ", str, log.Ldate|log.Ltime, 4)
	outInFile(username, "[Verbose] ", str, log.Ltime, 4)
}

func Info(username string, str string) {
	outInTerminal(username, "\x1B[1;32;1m[Info] ", str, log.Ldate|log.Ltime|log.Lshortfile, 4)
	outInFile(username, "[Info] ", str, log.Ltime|log.Lshortfile, 4)
}

func Debug(username string, str string, err error) {
	if err != nil {
		str += "，信息：" + err.Error()
	}
	outInTerminal(username, "\x1B[1;36;1m[Debug] ", str, log.Ldate|log.Ltime|log.Lshortfile, 4)
	outInFile(username, "[Debug] ", str, log.Ltime|log.Lshortfile, 4)
}

func Warn(username string, str string, err error) {
	if err != nil {
		str += "，信息：" + err.Error()
	}
	outInTerminal(username, "\x1B[1;33;1m[Warn] ", str, log.Ldate|log.Ltime|log.Lshortfile, 4)
	outInFile(username, "[Warn] ", str, log.Ltime|log.Lshortfile, 4)
}

func Error(username string, str string, errObj error) {
	if errObj != nil {
		str += "，信息：" + errObj.Error()
	}
	outInTerminal(username, "\x1B[1;31;1m[Error] ", str, log.Ldate|log.Ltime|log.Lshortfile, 4)
	outInFile(username, "[Error] ", str, log.Ltime|log.Lshortfile, 4)
}

func Assert(username string, str string, err error) {
	if err != nil {
		str += "，信息：" + err.Error()
	}
	outInTerminal(username, "\x1B[1;30;43m[Assert] ", str, log.Ldate|log.Ltime|log.Lshortfile, 4)
	outInFile(username, "[Assert] ", str, log.Ltime|log.Lshortfile, 4)
}

func GetEmptyErrorMessage() MessagedError {
	return MessagedError{
		HasInfo: false,
	}
}

func GetErrorMessage(code int, message string) MessagedError {
	return MessagedError{
		Code:    code,
		HasInfo: true,
		OutMessage: func(w http.ResponseWriter) {
			OnObjectResult(w, StringMessage{
				Code:    code,
				Message: message,
			})
		},
	}
}

func OnObjectResult(w http.ResponseWriter, object interface{}) {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	err := jsonEncoder.Encode(object)
	if err == nil {
		onResult(w, bf.Bytes())
	}
}

func OnStringResult(w http.ResponseWriter, str string) {
	onResult(w, []byte(str))
}

func onResult(w http.ResponseWriter, b []byte) {
	_, _ = w.Write(b)
}
