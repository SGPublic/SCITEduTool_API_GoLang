package main

import (
	"net/http"
	"os"

	"SCITEduTool/api"
	"SCITEduTool/base/Application"
	"SCITEduTool/unit/StdOutUnit"
)

func main() {
	Application.SetupWithConfig()
	RegisterAPI()
}

func RegisterAPI() {
	registerApi("/day", api.Day)
	registerApi("/hitokoto", api.Hitokoto)
	registerApi("/getKey", api.GetKey)
	registerApi("/login", api.Login)
	registerApi("/token", api.Token)
	registerApi("/springboard", api.Springboard)
	registerApi("/info", api.Info)
	registerApi("/table", api.Table)
	registerApi("/achieve", api.Achieve)
	registerApi("/exam", api.Exam)
	startService(addr)
}

const (
	basePattern = "/api"
	addr        = ":8000"
)

func registerApi(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(basePattern+pattern, handler)
}

func startService(addr string) {
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		StdOutUnit.Assert("", "服务启动失败", err)
		os.Exit(0)
	}
}
