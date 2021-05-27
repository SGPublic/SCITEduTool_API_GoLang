package api

import (
	"SCITEduTool/Application/manager"
	"SCITEduTool/Application/stdio"
	"net/http"
	"strconv"
	"time"
)

type BaseAPI struct {
	parameter         map[string]string
	OnObjectResult    func(object interface{})
	OnStandardMessage func(code int, message string)
	GetParameter      func(key string) string
}

func SetupAPI(w http.ResponseWriter, r *http.Request, parameterGet map[string]string) (BaseAPI, stdio.MessagedError) {
	parameter, sign, err := manager.SignManager.InsertParameter(r, parameterGet)
	if err.HasInfo {
		return BaseAPI{}, err
	}

	if !sign {
		return BaseAPI{}, stdio.GetErrorMessage(-403, "服务签名错误")
	}
	ts, intError := strconv.ParseInt(parameter["ts"], 10, 64)
	//IF !DEBUG
	if intError != nil {
		stdio.LogError("", "ts参数解析失败", intError)
		return BaseAPI{}, stdio.GetErrorMessage(-403, "请求错误")
	}
	timeNow := time.Now().Unix() - ts
	if timeNow > 600 || timeNow < -30 {
		return BaseAPI{}, stdio.GetErrorMessage(-408, "请求超时")
	}
	//ENDIF
	return BaseAPI{
		OnObjectResult: func(object interface{}) {
			stdio.OnObjectResult(w, object)
		},
		GetParameter: func(key string) string {
			return parameter[key]
		},
		OnStandardMessage: func(code int, message string) {
			stdio.OnObjectResult(w, struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				Code:    code,
				Message: message,
			})
		},
	}, stdio.GetEmptyErrorMessage()
}
