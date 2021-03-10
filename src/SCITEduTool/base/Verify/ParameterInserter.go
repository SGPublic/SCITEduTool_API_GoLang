package Verify

import (
	"SCITEduTool/base/LocalDebug"
	"SCITEduTool/unit/StdOutUnit"
	"net/http"
	"sort"
)

func InsertParameter(request *http.Request, parameter map[string]string) (map[string]string, StdOutUnit.MessagedError) {
	if parameter == nil {
		parameter = make(map[string]string)
	}
	if !LocalDebug.IsDebug() {
		parameter["ts"] = ""
		parameter["sign"] = ""
		parameter["platform"] = ""
		parameter["app_key"] = ""
	}
	parameterOut := make(map[string]string)
	parameterKeys := make([]string, 0)
	for key := range parameter {
		parameterKeys = append(parameterKeys, key)
	}
	sort.Strings(parameterKeys)
	for _, key := range parameterKeys {
		param := getParameter(request, key)
		if param != "" {
			parameterOut[key] = param
			continue
		}
		if parameter[key] == "" {
			username := getParameter(request, "username")
			if username == "" {
				username = "Unknown"
			}
			return nil, StdOutUnit.GetErrorMessage(-417, "参数缺失")
		}
		parameterOut[key] = parameter[key]
	}
	return parameterOut, StdOutUnit.MessagedError{}
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
