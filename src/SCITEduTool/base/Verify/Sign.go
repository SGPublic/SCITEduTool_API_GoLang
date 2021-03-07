package Verify

import (
	"SCITEduTool/base/LocalDebug"
	"SCITEduTool/unit/StdOutUnit"
	"crypto/md5"
	"encoding/hex"
)

func VerificationSign(parameter map[string]string) (bool, StdOutUnit.MessagedError) {
	if LocalDebug.IsDebug() {
		return true, StdOutUnit.GetEmptyErrorMessage()
	}
	parString := ""
	for key, value := range parameter {
		if parString != "" {
			parString += "&"
		}
		if key != "sign" {
			parString += key + "=" + value
		}
	}
	h := md5.New()
	h.Write([]byte(parString))
	sign := hex.EncodeToString(h.Sum(nil))
	if sign == parameter["sign"] {
		return true, StdOutUnit.GetEmptyErrorMessage()
	} else {
		return false, StdOutUnit.GetEmptyErrorMessage()
	}
}
