package Verify

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"sort"

	"SCITEduTool/base/LocalDebug"
	"SCITEduTool/manager/SignManager"
	"SCITEduTool/unit/StdOutUnit"
)

func InsertParameter(request *http.Request, parameter map[string]string) (map[string]string, bool, StdOutUnit.MessagedError) {
	if parameter == nil {
		parameter = make(map[string]string)
	}
	if !LocalDebug.IsDebug() {
		parameter["ts"] = ""
		parameter["sign"] = ""
		parameter["platform"] = "web"
		parameter["app_key"] = SignManager.GetDefaultAppSecretByPlatform("web")
	}
	parString := ""
	var parameterKeys []string
	for key := range parameter {
		parameterKeys = append(parameterKeys, key)
	}
	sort.Strings(parameterKeys)
	for _, key := range parameterKeys {
		param := getParameter(request, key)
		//if param == "@null" {
		//	return nil, false, StdOutUnit.GetErrorMessage(-417, "不支持的请求方式")
		//}
		if param != "" {
			parameter[key] = param
			if key != "sign" {
				if parString != "" {
					parString += "&"
				}
				parString += key + "=" + param
			}
			continue
		}
		if parameter[key] == "" {
			StdOutUnit.Info("", "{"+request.RequestURI+"} 请求参数缺失："+key)
			return nil, false, StdOutUnit.GetErrorMessage(-417, "参数缺失")
		}
	}

	if LocalDebug.IsDebug() {
		return parameter, true, StdOutUnit.GetEmptyErrorMessage()
	}
	appSecret := SignManager.GetAppSecretByAppKey(parameter["app_key"], parameter["platform"])
	if appSecret == "" {
		return nil, false, StdOutUnit.GetErrorMessage(-403, "应用密钥不存在")
	}

	h := md5.New()
	h.Write([]byte(parString + appSecret))
	sign := hex.EncodeToString(h.Sum(nil))
	if sign == parameter["sign"] {
		return parameter, true, StdOutUnit.GetEmptyErrorMessage()
	} else {
		StdOutUnit.Debug("", parString, nil)
		return nil, false, StdOutUnit.GetEmptyErrorMessage()
	}
}

func getParameter(request *http.Request, key string) string {
	value := request.PostFormValue(key)
	if value != "" {
		return value
	}
	value = request.FormValue(key)
	//if value != "" {
	//	return value
	//}
	//_ = request.ParseMultipartForm(1024 * 1024 * 5)
	//value = request.MultipartForm.Value[key][0]
	return value
}
