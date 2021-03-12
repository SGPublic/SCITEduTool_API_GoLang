package api

import (
	"SCITEduTool/base/LocalDebug"
	"SCITEduTool/base/Verify"
	"SCITEduTool/unit/StdOutUnit"
	"net/http"
	"strconv"
	"time"
)

type BaseAPI struct {
	parameter      map[string]string
	OnObjectResult func(object interface{})
	GetParameter   func(key string) string
}

func SetupAPI(w http.ResponseWriter, r *http.Request, parameterGet map[string]string) (BaseAPI, StdOutUnit.MessagedError) {
	parameter, err := Verify.InsertParameter(r, parameterGet)
	if err.HasInfo {
		return BaseAPI{}, err
	}
	sign, err := Verify.VerificationSign(parameter)
	if err.HasInfo {
		return BaseAPI{}, err
	}
	if !sign {
		return BaseAPI{}, StdOutUnit.GetErrorMessage(-403, "服务签名错误")
	}
	ts, intError := strconv.ParseInt(parameter["ts"], 10, 64)
	if !LocalDebug.IsDebug() {
		if intError != nil {
			StdOutUnit.Error("", "ts参数解析失败", intError)
			return BaseAPI{}, StdOutUnit.GetErrorMessage(-403, "请求错误")
		}
		timeNow := time.Now().Unix() - ts
		if timeNow > 600 || timeNow < -30 {
			return BaseAPI{}, StdOutUnit.GetErrorMessage(-408, "请求超时")
		}
	}
	return BaseAPI{
		OnObjectResult: func(object interface{}) {
			StdOutUnit.OnObjectResult(w, object)
		},
		GetParameter: func(key string) string {
			return parameter[key]
		},
	}, StdOutUnit.GetEmptyErrorMessage()
}
