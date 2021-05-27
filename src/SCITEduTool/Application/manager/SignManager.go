package manager

import (
	"SCITEduTool/Application/stdio"
	"SCITEduTool/Application/unit"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"net/http"
	"sort"
)

type signManager interface {
	InsertParameter(request *http.Request, parameter map[string]string) (map[string]string, bool, stdio.MessagedError)
	GetDefaultAppKey() string
	GetAppSecretByAppKey(appKey string, platform string) string
	GetDefaultAppSecretByPlatform(platform string) string
}

type signManagerImpl struct{}

var SignManager signManager = signManagerImpl{}

func (signManagerImpl signManagerImpl) InsertParameter(request *http.Request, parameter map[string]string) (map[string]string, bool, stdio.MessagedError) {
	if parameter == nil {
		parameter = make(map[string]string)
	}
	//IF !DEBUG
	parameter["ts"] = ""
	parameter["sign"] = ""
	parameter["platform"] = "web"
	parameter["app_key"] = SignManager.GetDefaultAppKey()
	//ENDIF
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
			stdio.LogInfo("", "{"+request.RequestURI+"} 请求参数缺失："+key)
			return nil, false, stdio.GetErrorMessage(-417, "参数缺失")
		}
	}

	//IF DEBUG
	//	return parameter, true, StdOutUnit.GetEmptyErrorMessage()
	//ENDIF
	appSecret := SignManager.GetAppSecretByAppKey(parameter["app_key"], parameter["platform"])
	if appSecret == "" {
		stdio.LogDebug("", parameter["app_key"], nil)
		stdio.LogDebug("", parameter["platform"], nil)
		return nil, false, stdio.GetErrorMessage(-403, "应用密钥不存在")
	}

	h := md5.New()
	h.Write([]byte(parString + appSecret))
	sign := hex.EncodeToString(h.Sum(nil))
	if sign == parameter["sign"] {
		return parameter, true, stdio.GetEmptyErrorMessage()
	} else {
		stdio.LogDebug("", parString, nil)
		return nil, false, stdio.GetEmptyErrorMessage()
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

func (signManagerImpl signManagerImpl) GetDefaultAppKey() string {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return ""
	}
	state, err := tx.Prepare("select `app_key` from `sign_keys` where `platform`='web' order by `build` desc limit 1")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return ""
	}
	rows := state.QueryRow()
	appKey := ""
	err = rows.Scan(&appKey)
	if err == nil {
		tx.Commit()
		return appKey
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return ""
	} else {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return ""
	}
}

func (signManagerImpl signManagerImpl) GetAppSecretByAppKey(appKey string, platform string) string {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return ""
	}
	state, err := tx.Prepare("select `app_secret` from `sign_keys` where `app_key`=? and `platform`=?")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return ""
	}
	rows := state.QueryRow(appKey, platform)
	secret := ""
	err = rows.Scan(&secret)
	if err == nil {
		tx.Commit()
		return secret
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return ""
	} else {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return ""
	}
}

func (signManagerImpl signManagerImpl) GetDefaultAppSecretByPlatform(platform string) string {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return ""
	}
	state, err := tx.Prepare("select `app_secret` from `sign_keys` where `platform`=? order by `build` desc limit 1")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return ""
	}
	rows := state.QueryRow(platform)
	secret := ""
	err = rows.Scan(&secret)
	if err == nil {
		tx.Commit()
		return secret
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return ""
	} else {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return ""
	}
}
