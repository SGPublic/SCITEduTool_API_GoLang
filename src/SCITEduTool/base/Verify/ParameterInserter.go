package Verify

import (
	"net/http"

	"SCITEduTool/base/LocalDebug"
	"SCITEduTool/manager/SignManager"
	"SCITEduTool/unit/StdOutUnit"
)

func InsertParameter(request *http.Request, parameter map[string]string) (map[string]string, StdOutUnit.MessagedError) {
	if parameter == nil {
		parameter = make(map[string]string)
	}
	if !LocalDebug.IsDebug() {
		parameter["ts"] = ""
		parameter["sign"] = ""
		parameter["platform"] = "web"
		parameter["app_key"] = SignManager.GetDefaultAppSecretByPlatform("web")
	}
	for key := range parameter {
		param := getParameter(request, key)
		if param != "" {
			parameter[key] = param
			continue
		}
		if parameter[key] == "" {
			StdOutUnit.Info("", "{"+request.RequestURI+"} 请求参数缺失："+key)
			return nil, StdOutUnit.GetErrorMessage(-417, "参数缺失")
		}
		parameter[key] = parameter[key]
	}
	return parameter, StdOutUnit.MessagedError{}
}

func getParameter(request *http.Request, key string) string {
	value := request.PostFormValue(key)
	if value != "" {
		return value
	}
	if LocalDebug.IsDebug() {
		return request.FormValue(key)
	}
	return ""
}
