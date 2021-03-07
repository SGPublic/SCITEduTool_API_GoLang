package StdOutUnit

import (
	"SCITEduTool/base/LocalDebug"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type MessagedError struct {
	Message    StringMessage
	Info       StringMessage
	HasInfo    bool
	OutLog func(username string)
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
	doOut(os.Stdout, prefix, str + "\x1b[0m", flag, calldepth)
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
	file, err := os.OpenFile(path, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	doOut(io.Writer(file), prefix, str, flag, calldepth)
}

func doOut(writer io.Writer, prefix string, str string, flag int, calldepth int) {
	MyLog := log.New(writer, prefix, flag)
	_ = MyLog.Output(calldepth, str)
}

type VerboseLog struct { }
func (verbose VerboseLog) String(username string, str string) {
	outInTerminal(username, "\x1B[1;37;1m[Verbose] ", str, log.Ldate | log.Ltime, 4)
	outInFile(username, "[Verbose] ", str, log.Ltime, 4)
}
func (verbose VerboseLog) Object(username string, obj interface{}) {
	str, _ := json.Marshal(obj)
	outInTerminal(username, "\x1B[1;37;1m[Verbose] ", string(str), log.Ldate | log.Ltime, 4)
	outInFile(username, "[Verbose] ", string(str), log.Ltime, 4)
}
var Verbose = new(VerboseLog)

type InfoLog struct { }
func (info InfoLog) String(username string, str string) {
	outInTerminal(username, "\x1B[1;32;1m[Info] ", str, log.Ldate | log.Ltime | log.Lshortfile, 4)
	outInFile(username, "[Info] ", str, log.Ltime | log.Lshortfile, 4)
}
func (info InfoLog) Object(username string, obj interface{}) {
	str, _ := json.Marshal(obj)
	outInTerminal(username, "\x1B[1;32;1m[Info] ", string(str), log.Ldate | log.Ltime | log.Lshortfile, 4)
	outInFile(username, "[Info] ", string(str), log.Ltime | log.Lshortfile, 4)
}
var Info = new(InfoLog)

type DebugLog struct { }
func (debug DebugLog) String(username string, str string) {
	outInTerminal(username, "\x1B[1;36;1m[Debug] ", str, log.Ldate | log.Ltime | log.Lshortfile, 4)
	outInFile(username, "[Debug] ", str, log.Ltime | log.Lshortfile, 4)
}
func (debug DebugLog) Object(username string, obj interface{}) {
	str, _ := json.Marshal(obj)
	outInTerminal(username, "\x1B[1;36;1m[Debug] ", string(str), log.Ldate | log.Ltime | log.Lshortfile, 4)
	outInFile(username, "[Debug] ", string(str), log.Ltime | log.Lshortfile, 4)
}
var Debug = new(DebugLog)

type WarnLog struct { }
func (warn WarnLog) String(username string, str string) {
	outInTerminal(username, "\x1B[1;33;1m[Warn] ", str, log.Ldate | log.Ltime | log.Lshortfile, 4)
	outInFile(username, "[Warn] ", str, log.Ltime | log.Lshortfile, 4)
}
func (warn WarnLog) Object(username string, obj interface{}) {
	str, _ := json.Marshal(obj)
	outInTerminal(username, "\x1B[1;33;1m[Warn] ", string(str), log.Ldate | log.Ltime | log.Lshortfile, 4)
	outInFile(username, "[Warn] ", string(str), log.Ltime | log.Lshortfile, 4)
}
var Warn = new(WarnLog)

type ErrorLog struct { }
func (err ErrorLog) String(username string, str string) {
	outInTerminal(username, "\x1B[1;31;1m[Error] ", str, log.Ldate | log.Ltime | log.Lshortfile, 4)
	outInFile(username, "[Error] ", str, log.Ltime | log.Lshortfile, 4)
}
func (err ErrorLog) Object(username string, obj interface{}) {
	str, _ := json.Marshal(obj)
	outInTerminal(username, "\x1B[1;31;1m[Error] ", string(str), log.Ldate | log.Ltime | log.Lshortfile, 4)
	outInFile(username, "[Error] ", string(str), log.Ltime | log.Lshortfile, 4)
}
var Error = new(ErrorLog)

type AssertLog struct { }
func (assert AssertLog) String(username string, str string) {
	outInTerminal(username, "\x1B[1;30;43m[Assert] ", str, log.Ldate | log.Ltime | log.Lshortfile, 4)
	outInFile(username, "[Assert] ", str, log.Ltime | log.Lshortfile, 4)
}
func (assert AssertLog) Object(username string, obj interface{}) {
	str, _ := json.Marshal(obj)
	outInTerminal(username, "\x1B[1;30;43m[Assert] ", string(str), log.Ldate | log.Ltime | log.Lshortfile, 4)
	outInFile(username, "[Assert] ", string(str), log.Ltime | log.Lshortfile, 4)
}
var Assert = new(AssertLog)

func GetEmptyErrorMessage() MessagedError {
	return MessagedError {
		HasInfo: false,
	}
}

func GetErrorMessage(code int, message string) MessagedError {
	messagedInfo := StringMessage {
		Code:    code,
		Message: message,
	}
	return MessagedError {
		HasInfo: true,
		OutLog: func(username string) {
			str, _ := json.Marshal(messagedInfo)
			outInTerminal(username, "\x1B[1;31;1m[Error] ", string(str), log.Ldate | log.Ltime | log.Lshortfile, 4)
			outInFile(username, "[Error] ", string(str), log.Ltime | log.Lshortfile, 4)
		},
		OutMessage: func(w http.ResponseWriter) {
			OnObjectResult(w, messagedInfo)
		},
	}
}

func OnObjectResult(w http.ResponseWriter, object interface{}) {
	result, _ := json.Marshal(object)
	onResult(w, result)
}

func OnStringResult(w http.ResponseWriter, str string) {
	onResult(w, []byte(str))
}

func onResult(w http.ResponseWriter, b []byte) {
	_, _ = w.Write(b)
}
