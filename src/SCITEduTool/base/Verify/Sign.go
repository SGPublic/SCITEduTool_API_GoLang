package Verify

import (
	"SCITEduTool/base/LocalDebug"
	"SCITEduTool/manager/SignManager"
	"SCITEduTool/unit/StdOutUnit"
	"crypto/md5"
	"encoding/hex"
	"sort"
)

func VerificationSign(parameter map[string]string) (bool, StdOutUnit.MessagedError) {
	if LocalDebug.IsDebug() {
		return true, StdOutUnit.GetEmptyErrorMessage()
	}
	parString := ""
	var parameterKeys []string
	for key := range parameter {
		parameterKeys = append(parameterKeys, key)
	}
	sort.Strings(parameterKeys)
	for _, key := range parameterKeys {
		if key == "sign" {
			continue
		}
		if parString != "" {
			parString += "&"
		}
		parString += key + "=" + parameter[key]
	}

	appSecret := SignManager.GetAppSecretByAppKey(parameter["app_key"], parameter["platform"])
	if appSecret == "" {
		return false, StdOutUnit.GetErrorMessage(-403, "应用密钥不存在")
	}

	h := md5.New()
	h.Write([]byte(parString + appSecret))
	sign := hex.EncodeToString(h.Sum(nil))
	if sign == parameter["sign"] {
		return true, StdOutUnit.GetEmptyErrorMessage()
	} else {
		StdOutUnit.Debug("", parString, nil)
		return false, StdOutUnit.GetEmptyErrorMessage()
	}
}
